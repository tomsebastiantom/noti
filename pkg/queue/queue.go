package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"getnoti.com/pkg/circuitbreaker"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/workerpool"
	"github.com/streadway/amqp"
)

type Message struct {
    Body []byte
    Headers map[string]interface{}
    Timestamp time.Time
}

type Queue interface {
    GetOrCreateChannel(channelName string) (*amqp.Channel, error)
    Publish(ctx context.Context, channelName, routingKey string, msg Message) error
    Consume(ctx context.Context, channelName, queueName string) (<-chan Message, error)
    DeclareQueue(ctx context.Context, channelName, queueName string, durable, autoDelete, exclusive bool) error
    InitializeConsumer(ctx context.Context, channelName, queueName string, handler func(Message), workerPool *workerpool.WorkerPool) error
    Close() error
    IsHealthy() bool
    Ping() error
}

type AMQPQueue struct {
    conn           *amqp.Connection
    channels       map[string]*amqp.Channel
    config         Config
    log            logger.Logger
    circuitBreaker *circuitbreaker.CircuitBreaker
    mu             sync.RWMutex
    state          int32 // 0: connected, 1: disconnected, 2: closed
    metrics        *QueueMetrics
    metricsMu      sync.RWMutex
}

type QueueMetrics struct {
    MessagesPublished int64
    MessagesConsumed  int64
    PublishErrors     int64
    ConsumeErrors     int64
    ConnectionCount   int64
    LastActivity      time.Time
    CreatedAt         time.Time
}

type Config struct {
    URL               string
    ReconnectInterval time.Duration
    MaxReconnectAttempts int
    HeartbeatInterval time.Duration
}

type QueueManager struct {
    queues   map[string]Queue
    config   Config
    log      logger.Logger
    mu       sync.RWMutex
    shutdown chan struct{}
    once     sync.Once
}

// NewQueueManager creates a new queue manager
func NewQueueManager(config Config, log logger.Logger) *QueueManager {
    // Set defaults
    if config.ReconnectInterval == 0 {
        config.ReconnectInterval = 5 * time.Second
    }
    if config.MaxReconnectAttempts == 0 {
        config.MaxReconnectAttempts = 10
    }
    if config.HeartbeatInterval == 0 {
        config.HeartbeatInterval = 10 * time.Second
    }
    
    return &QueueManager{
        queues:   make(map[string]Queue),
        config:   config,
        log:      log,
        shutdown: make(chan struct{}),
    }
}

// NewAMQPQueue creates a new AMQP queue with enhanced connection management
func NewAMQPQueue(config Config, log logger.Logger) (*AMQPQueue, error) {
    conn, err := amqp.Dial(config.URL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to AMQP: %w", err)
    }
    
    q := &AMQPQueue{
        conn:           conn,
        channels:       make(map[string]*amqp.Channel),
        config:         config,
        log:            log.With(logger.Field{Key: "component", Value: "amqp_queue"}),
        circuitBreaker: circuitbreaker.NewCircuitBreaker(5, 3, 60),
        state:          0, // connected
        metrics: &QueueMetrics{
            CreatedAt: time.Now(),
        },
    }
    
    // Start connection monitoring
    go q.connectionMonitor()
    
    q.log.Info("AMQP queue initialized successfully",
        logger.Field{Key: "url", Value: config.URL})
    
    return q, nil
}

func (q *AMQPQueue) GetOrCreateChannel(channelName string) (*amqp.Channel, error) {
    q.mu.RLock()
    ch, exists := q.channels[channelName]
    q.mu.RUnlock()
    
    // ✅ Fixed: Remove IsClosed() check for channels
    if exists && ch != nil {
        // Test if channel is still usable by trying a simple operation
        if err := ch.ExchangeDeclarePassive("", "direct", true, false, false, false, nil); err == nil {
            return ch, nil
        }
        // If test fails, the channel is likely closed, so we'll create a new one
    }
    
    q.mu.Lock()
    defer q.mu.Unlock()
    
    // ✅ Fixed: Double-check locking without IsClosed()
    if ch, exists = q.channels[channelName]; exists && ch != nil {
        // Test again under write lock
        if err := ch.ExchangeDeclarePassive("", "direct", true, false, false, false, nil); err == nil {
            return ch, nil
        }
        // Channel is dead, remove it
        delete(q.channels, channelName)
    }
    
    // Check connection state
    if atomic.LoadInt32(&q.state) != 0 {
        return nil, fmt.Errorf("queue connection is not available")
    }
    
    newCh, err := q.conn.Channel()
    if err != nil {
        q.log.Error("Failed to create channel",
            logger.Field{Key: "channel_name", Value: channelName},
            logger.Field{Key: "error", Value: err.Error()})
        return nil, fmt.Errorf("failed to create channel %s: %w", channelName, err)
    }
    
    q.channels[channelName] = newCh
    q.log.Debug("Created new channel",
        logger.Field{Key: "channel_name", Value: channelName})
    
    return newCh, nil
}

// Publish publishes a message with retry logic and monitoring
func (q *AMQPQueue) Publish(ctx context.Context, channelName, routingKey string, msg Message) error {
    return q.circuitBreaker.Execute(func() error {
        ch, err := q.GetOrCreateChannel(channelName)
        if err != nil {
            return err
        }
        
        // Add timestamp if not present
        if msg.Timestamp.IsZero() {
            msg.Timestamp = time.Now()
        }
        
        // Prepare headers
        headers := make(amqp.Table)
        for k, v := range msg.Headers {
            headers[k] = v
        }
        headers["timestamp"] = msg.Timestamp
        
        publishing := amqp.Publishing{
            ContentType: "application/json",
            Body:        msg.Body,
            Headers:     headers,
            Timestamp:   msg.Timestamp,
            MessageId:   fmt.Sprintf("%d", time.Now().UnixNano()),
        }
        
        err = ch.Publish(
            "",         // exchange
            routingKey, // routing key
            false,      // mandatory
            false,      // immediate
            publishing,
        )
        
        if err != nil {
            atomic.AddInt64(&q.metrics.PublishErrors, 1)
            q.log.Error("Failed to publish message",
                logger.Field{Key: "routing_key", Value: routingKey},
                logger.Field{Key: "error", Value: err.Error()})
            return err
        }
        
        atomic.AddInt64(&q.metrics.MessagesPublished, 1)
        q.updateLastActivity()
        
        q.log.Debug("Message published successfully",
            logger.Field{Key: "routing_key", Value: routingKey},
            logger.Field{Key: "message_size", Value: len(msg.Body)})
        
        return nil
    })
}

// Consume consumes messages with proper error handling
func (q *AMQPQueue) Consume(ctx context.Context, channelName, queueName string) (<-chan Message, error) {
    var messages <-chan Message
    
    err := q.circuitBreaker.Execute(func() error {
        ch, err := q.GetOrCreateChannel(channelName)
        if err != nil {
            return err
        }
        
        deliveries, err := ch.Consume(
            queueName, // queue
            "",        // consumer
            false,     // auto-ack (manual ack for reliability)
            false,     // exclusive
            false,     // no-local
            false,     // no-wait
            nil,       // args
        )
        if err != nil {
            atomic.AddInt64(&q.metrics.ConsumeErrors, 1)
            return fmt.Errorf("failed to start consuming from queue %s: %w", queueName, err)
        }
        
        msgChan := make(chan Message, 100) // Buffered channel
        go func() {
            defer close(msgChan)
            
            for {
                select {
                case d, ok := <-deliveries:
                    if !ok {
                        q.log.Warn("Delivery channel closed",
                            logger.Field{Key: "queue_name", Value: queueName})
                        return
                    }
                    
                    msg := Message{
                        Body:      d.Body,
                        Headers:   make(map[string]interface{}),
                        Timestamp: d.Timestamp,
                    }
                    
                    // Copy headers
                    for k, v := range d.Headers {
                        msg.Headers[k] = v
                    }
                    
                    select {
                    case msgChan <- msg:
                        atomic.AddInt64(&q.metrics.MessagesConsumed, 1)
                        q.updateLastActivity()
                        d.Ack(false) // Acknowledge message
                    case <-ctx.Done():
                        d.Nack(false, true) // Negative ack and requeue
                        return
                    }
                    
                case <-ctx.Done():
                    q.log.Info("Consumer context cancelled",
                        logger.Field{Key: "queue_name", Value: queueName})
                    return
                }
            }
        }()
        
        messages = msgChan
        return nil
    })
    
    return messages, err
}

// DeclareQueue declares a queue with proper error handling
func (q *AMQPQueue) DeclareQueue(ctx context.Context, channelName, queueName string, durable, autoDelete, exclusive bool) error {
    return q.circuitBreaker.Execute(func() error {
        ch, err := q.GetOrCreateChannel(channelName)
        if err != nil {
            return err
        }
        
        _, err = ch.QueueDeclare(
            queueName,
            durable,
            autoDelete,
            exclusive,
            false, // no-wait
            nil,   // arguments
        )
        
        if err != nil {
            q.log.Error("Failed to declare queue",
                logger.Field{Key: "queue_name", Value: queueName},
                logger.Field{Key: "error", Value: err.Error()})
            return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
        }
        
        q.log.Info("Queue declared successfully",
            logger.Field{Key: "queue_name", Value: queueName},
            logger.Field{Key: "durable", Value: durable},
            logger.Field{Key: "auto_delete", Value: autoDelete},
            logger.Field{Key: "exclusive", Value: exclusive})
        
        return nil
    })
}

// InitializeConsumer sets up a consumer with worker pool integration
func (q *AMQPQueue) InitializeConsumer(ctx context.Context, channelName, queueName string, handler func(Message), workerPool *workerpool.WorkerPool) error {
    return q.circuitBreaker.Execute(func() error {
        ch, err := q.GetOrCreateChannel(channelName)
        if err != nil {
            return err
        }
        
        msgs, err := ch.Consume(
            queueName, // queue
            "",        // consumer
            false,     // auto-ack (manual acknowledgment for reliability)
            false,     // exclusive
            false,     // no-local
            false,     // no-wait
            nil,       // args
        )
        if err != nil {
            return fmt.Errorf("failed to start consumer for queue %s: %w", queueName, err)
        }
        
        q.log.Info("Consumer initialized",
            logger.Field{Key: "queue_name", Value: queueName})
        
        go func() {
            for {
                select {
                case msg, ok := <-msgs:
                    if !ok {
                        q.log.Warn("Channel closed for queue",
                            logger.Field{Key: "queue", Value: queueName})
                        return
                    }
                    
                    // Create a job for the worker pool
                    job := &ConsumerJob{
                        message: Message{
                            Body:      msg.Body,
                            Headers:   make(map[string]interface{}),
                            Timestamp: msg.Timestamp,
                        },
                        handler: handler,
                        ack: func() {
                            msg.Ack(false)
                        },
                        nack: func() {
                            msg.Nack(false, true)
                        },
                        logger: q.log,
                    }
                    
                    // Copy headers
                    for k, v := range msg.Headers {
                        job.message.Headers[k] = v
                    }
                    
                    // Submit the job to the worker pool
                    err := workerPool.Submit(job)
                    if err != nil {
                        q.log.Error("Failed to submit job to worker pool",
                            logger.Field{Key: "error", Value: err.Error()},
                            logger.Field{Key: "queue", Value: queueName})
                        msg.Nack(false, true) // Negative acknowledge and requeue the message
                    } else {
                        atomic.AddInt64(&q.metrics.MessagesConsumed, 1)
                        q.updateLastActivity()
                    }
                    
                case <-ctx.Done():
                    q.log.Info("Context cancelled for consumer",
                        logger.Field{Key: "queue", Value: queueName})
                    return
                }
            }
        }()
        
        return nil
    })
}

// connectionMonitor monitors connection health and handles reconnection
func (q *AMQPQueue) connectionMonitor() {
    for {
        reason, ok := <-q.conn.NotifyClose(make(chan *amqp.Error))
        if !ok {
            q.log.Info("Connection monitor stopped")
            break
        }
        
        atomic.StoreInt32(&q.state, 1) // disconnected
        q.log.Warn("Connection lost, attempting to reconnect",
            logger.Field{Key: "reason", Value: reason.Error()})
        
        if err := q.reconnectWithBackoff(); err != nil {
            q.log.Error("Failed to reconnect after all attempts",
                logger.Field{Key: "error", Value: err.Error()})
            atomic.StoreInt32(&q.state, 2) // closed
            return
        }
        
        atomic.StoreInt32(&q.state, 0) // connected
        q.log.Info("Reconnected successfully")
    }
}

// reconnectWithBackoff implements exponential backoff reconnection
func (q *AMQPQueue) reconnectWithBackoff() error {
    backoff := q.config.ReconnectInterval
    maxBackoff := 60 * time.Second
    
    for attempt := 1; attempt <= q.config.MaxReconnectAttempts; attempt++ {
        q.log.Info("Attempting to reconnect",
            logger.Field{Key: "attempt", Value: attempt},
            logger.Field{Key: "max_attempts", Value: q.config.MaxReconnectAttempts})
        
        conn, err := amqp.Dial(q.config.URL)
        if err == nil {
            q.mu.Lock()
            q.conn = conn
            q.channels = make(map[string]*amqp.Channel) // Clear old channels
            atomic.AddInt64(&q.metrics.ConnectionCount, 1)
            q.mu.Unlock()
            
            return nil
        }
        
        q.log.Warn("Reconnection attempt failed",
            logger.Field{Key: "attempt", Value: attempt},
            logger.Field{Key: "error", Value: err.Error()},
            logger.Field{Key: "retry_in", Value: backoff.String()})
        
        time.Sleep(backoff)
        
        // Exponential backoff with jitter
        backoff = time.Duration(float64(backoff) * 1.5)
        if backoff > maxBackoff {
            backoff = maxBackoff
        }
    }
    
    return fmt.Errorf("failed to reconnect after %d attempts", q.config.MaxReconnectAttempts)
}

// updateLastActivity updates the last activity timestamp
func (q *AMQPQueue) updateLastActivity() {
    q.metricsMu.Lock()
    q.metrics.LastActivity = time.Now()
    q.metricsMu.Unlock()
}

// IsHealthy checks if the queue is healthy
func (q *AMQPQueue) IsHealthy() bool {
    // Check connection state
    if atomic.LoadInt32(&q.state) != 0 {
        return false
    }
    
    // Check if connection exists and is not closed
    q.mu.RLock()
    conn := q.conn
    q.mu.RUnlock()
    
    if conn == nil || conn.IsClosed() {
        return false
    }
    
    return true
}

// Ping performs a health check
func (q *AMQPQueue) Ping() error {
    if atomic.LoadInt32(&q.state) != 0 {
        return fmt.Errorf("queue is not connected")
    }
    
    q.mu.RLock()
    conn := q.conn
    q.mu.RUnlock()
    
    if conn == nil {
        return fmt.Errorf("connection is nil")
    }
    
    if conn.IsClosed() {
        return fmt.Errorf("connection is closed")
    }
    
    // Try to create and close a channel as a ping
    ch, err := conn.Channel()
    if err != nil {
        return fmt.Errorf("failed to create test channel: %w", err)
    }
    
    return ch.Close()
}

// Close closes the queue and all its channels
func (q *AMQPQueue) Close() error {
    atomic.StoreInt32(&q.state, 2) // closed
    
    q.mu.Lock()
    defer q.mu.Unlock()
    
    // Close all channels
    for name, ch := range q.channels {
        if err := ch.Close(); err != nil {
            q.log.Error("Failed to close channel",
                logger.Field{Key: "channel_name", Value: name},
                logger.Field{Key: "error", Value: err.Error()})
        }
    }
    
    // Close connection
    if q.conn != nil && !q.conn.IsClosed() {
        if err := q.conn.Close(); err != nil {
            q.log.Error("Failed to close connection",
                logger.Field{Key: "error", Value: err.Error()})
            return err
        }
    }
    
    q.log.Info("AMQP queue closed successfully")
    return nil
}

// GetMetrics returns queue metrics
func (q *AMQPQueue) GetMetrics() QueueMetrics {
    q.metricsMu.RLock()
    defer q.metricsMu.RUnlock()
    
    metrics := *q.metrics
    metrics.MessagesPublished = atomic.LoadInt64(&q.metrics.MessagesPublished)
    metrics.MessagesConsumed = atomic.LoadInt64(&q.metrics.MessagesConsumed)
    metrics.PublishErrors = atomic.LoadInt64(&q.metrics.PublishErrors)
    metrics.ConsumeErrors = atomic.LoadInt64(&q.metrics.ConsumeErrors)
    metrics.ConnectionCount = atomic.LoadInt64(&q.metrics.ConnectionCount)
    
    return metrics
}

// QueueManager methods

// GetOrCreateQueue gets an existing queue or creates a new one
func (qm *QueueManager) GetOrCreateQueue(name string) (Queue, error) {
    qm.mu.RLock()
    q, exists := qm.queues[name]
    qm.mu.RUnlock()
    
    if exists {
        return q, nil
    }
    
    qm.mu.Lock()
    defer qm.mu.Unlock()
    
    // Double-check locking
    if q, exists = qm.queues[name]; exists {
        return q, nil
    }
    
    q, err := NewAMQPQueue(qm.config, qm.log)
    if err != nil {
        return nil, fmt.Errorf("failed to create queue %s: %w", name, err)
    }
    
    qm.queues[name] = q
    qm.log.Info("Created new queue",
        logger.Field{Key: "queue_name", Value: name})
    
    return q, nil
}

// Close gracefully closes all queue connections
func (qm *QueueManager) Close() error {
    qm.once.Do(func() {
        close(qm.shutdown)
    })
    
    qm.mu.Lock()
    defer qm.mu.Unlock()
    
    qm.log.Info("Closing all queue connections",
        logger.Field{Key: "queue_count", Value: len(qm.queues)})
    
    var closeErrors []error
    
    for name, queue := range qm.queues {
        qm.log.Debug("Closing queue",
            logger.Field{Key: "queue_name", Value: name})
        
        if err := queue.Close(); err != nil {
            qm.log.Error("Failed to close queue",
                logger.Field{Key: "queue_name", Value: name},
                logger.Field{Key: "error", Value: err.Error()})
            closeErrors = append(closeErrors, err)
        } else {
            qm.log.Debug("Queue closed successfully",
                logger.Field{Key: "queue_name", Value: name})
        }
        delete(qm.queues, name)
    }
    
    if len(closeErrors) > 0 {
        qm.log.Error("Queue manager close completed with errors",
            logger.Field{Key: "error_count", Value: len(closeErrors)})
        return fmt.Errorf("failed to close %d queues", len(closeErrors))
    }
    
    qm.log.Info("All queue connections closed successfully")
    return nil
}

// IsHealthy checks if all queues are healthy
func (qm *QueueManager) IsHealthy() bool {
    qm.mu.RLock()
    defer qm.mu.RUnlock()
    
    for name, queue := range qm.queues {
        if !queue.IsHealthy() {
            qm.log.Warn("Unhealthy queue detected",
                logger.Field{Key: "queue_name", Value: name})
            return false
        }
    }
    return true
}

// GetHealthStatus returns detailed health status of all queues
func (qm *QueueManager) GetHealthStatus() map[string]bool {
    qm.mu.RLock()
    defer qm.mu.RUnlock()
    
    status := make(map[string]bool)
    for name, queue := range qm.queues {
        status[name] = queue.IsHealthy()
    }
    return status
}

// ConsumerJob implements the Job interface for the worker pool
type ConsumerJob struct {
    message Message
    handler func(Message)
    ack     func()
    nack    func()
    logger  logger.Logger
}

// Process processes the consumer job
func (j *ConsumerJob) Process(ctx context.Context) error {
    defer func() {
        if r := recover(); r != nil {
            j.logger.Error("Consumer job panic recovered",
                logger.Field{Key: "panic", Value: fmt.Sprintf("%v", r)})
            j.nack() // Negative acknowledge on panic
        }
    }()
    
    // Process the message with timeout
    done := make(chan struct{})
    var processingErr error
    
    go func() {
        defer func() {
            if r := recover(); r != nil {
                processingErr = fmt.Errorf("handler panic: %v", r)
            }
            close(done)
        }()
        
        j.handler(j.message)
    }()
    
    select {
    case <-done:
        if processingErr != nil {
            j.nack()
            return processingErr
        }
        j.ack()
        return nil
    case <-ctx.Done():
        j.nack()
        return ctx.Err()
    }
}