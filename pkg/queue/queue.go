package queue

import (
	"context"
	"sync"
	"time"

	"getnoti.com/pkg/circuitbreaker"
	"getnoti.com/pkg/logger"
	"github.com/streadway/amqp"
)

type Message struct {
	Body []byte
}

type Queue interface {
	GetOrCreateChannel(channelName string) (*amqp.Channel, error)
	Publish(ctx context.Context, channelName, routingKey string, msg Message) error
	Consume(ctx context.Context, channelName, queueName string) (<-chan Message, error)
	DeclareQueue(ctx context.Context, channelName, queueName string, durable, autoDelete, exclusive bool) error
	InitializeConsumer(ctx context.Context, channelName, queueName string, handler func(Message)) error
	Close() error
}

type AMQPQueue struct {
	conn           *amqp.Connection
	channels       map[string]*amqp.Channel
	config         Config
	log            logger.Interface
	circuitBreaker *circuitbreaker.CircuitBreaker
	mu             sync.RWMutex
}

type Config struct {
	URL               string
	ReconnectInterval time.Duration
}

type QueueManager struct {
	queues map[string]Queue
	config Config
	log    logger.Interface
	mu     sync.RWMutex
}

func NewQueueManager(config Config, log logger.Interface) *QueueManager {
	return &QueueManager{
		queues: make(map[string]Queue),
		config: config,
		log:    log,
	}
}

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
		return nil, err
	}

	qm.queues[name] = q
	return q, nil
}

func NewAMQPQueue(config Config, log logger.Interface) (*AMQPQueue, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	q := &AMQPQueue{
		conn:           conn,
		channels:       make(map[string]*amqp.Channel),
		config:         config,
		log:            log,
		circuitBreaker: circuitbreaker.NewCircuitBreaker(5, 3, 60),
	}

	go q.reconnectLoop()

	return q, nil
}

func (q *AMQPQueue) GetOrCreateChannel(channelName string) (*amqp.Channel, error) {
	q.mu.RLock()
	ch, exists := q.channels[channelName]
	q.mu.RUnlock()

	if exists {
		return ch, nil
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// Double-check locking
	if ch, exists = q.channels[channelName]; exists {
		return ch, nil
	}

	newCh, err := q.conn.Channel()
	if err != nil {
		return nil, err
	}

	q.channels[channelName] = newCh
	return newCh, nil
}

func (q *AMQPQueue) Publish(ctx context.Context, channelName, routingKey string, msg Message) error {
	return q.circuitBreaker.Execute(func() error {
		ch, err := q.GetOrCreateChannel(channelName)
		if err != nil {
			return err
		}

		return ch.Publish(
			"",         // exchange
			routingKey, // routing key
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        msg.Body,
			})
	})
}

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
			true,      // auto-ack
			false,     // exclusive
			false,     // no-local
			false,     // no-wait
			nil,       // args
		)
		if err != nil {
			return err
		}

		msgChan := make(chan Message)
		go func() {
			for d := range deliveries {
				select {
				case msgChan <- Message{Body: d.Body}:
				case <-ctx.Done():
					return
				}
			}
		}()

		messages = msgChan
		return nil
	})
	return messages, err
}

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
		return err
	})
}

func (q *AMQPQueue) reconnectLoop() {
	for {
		reason, ok := <-q.conn.NotifyClose(make(chan *amqp.Error))
		if !ok {
			q.log.Info("Connection closed. Exiting reconnect loop.")
			break
		}
		q.log.Warn("Connection closed: %s. Attempting to reconnect...", reason)

		for {
			err := q.connect()
			if err == nil {
				q.log.Info("Reconnected successfully")
				break
			}
			q.log.Error("Failed to reconnect: %s. Retrying in %s", err, q.config.ReconnectInterval)
			time.Sleep(q.config.ReconnectInterval)
		}
	}
}

func (q *AMQPQueue) connect() error {
	conn, err := amqp.Dial(q.config.URL)
	if err != nil {
		return err
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.conn = conn
	q.channels = make(map[string]*amqp.Channel)

	return nil
}
func (q *AMQPQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, ch := range q.channels {
		if err := ch.Close(); err != nil {
			q.log.Error("Failed to close channel: %s", err)
		}
	}

	return q.conn.Close()
}

func (q *AMQPQueue) InitializeConsumer(ctx context.Context, channelName, queueName string, handler func(Message)) error {
    return q.circuitBreaker.Execute(func() error {
        ch, err := q.GetOrCreateChannel(channelName)
        if err != nil {
            return err
        }

        msgs, err := ch.Consume(
            queueName, // queue
            "",        // consumer
            true,      // auto-ack
            false,     // exclusive
            false,     // no-local
            false,     // no-wait
            nil,       // args
        )
        if err != nil {
            return err
        }

        go func() {
            for {
                select {
                case msg, ok := <-msgs:
                    if !ok {
                        q.log.Warn("Channel closed for queue: %s", queueName)
                        return
                    }
                    handler(Message{Body: msg.Body})
                case <-ctx.Done():
                    q.log.Info("Context cancelled for consumer on queue: %s", queueName)
                    return
                }
            }
        }()

        return nil
    })
}
