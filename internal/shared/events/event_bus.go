package events

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"getnoti.com/pkg/db"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/workerpool"
)

// EventHandler defines the function signature for event handlers
type EventHandler func(context.Context, DomainEvent) error

// EventBus defines the interface for event publishing and subscribing
type EventBus interface {
	// Synchronous event publishing (within same transaction)
	PublishSync(ctx context.Context, event DomainEvent) error
	
	// Asynchronous event publishing (via worker pool)
	PublishAsync(ctx context.Context, event DomainEvent) error
	
	// Subscribe to events (handlers run synchronously)
	Subscribe(eventType string, handler EventHandler) error
	
	// Subscribe to async events (handlers run asynchronously via worker pool)
	SubscribeAsync(eventType string, handler EventHandler) error
	
	// Lifecycle management
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// HybridEventBus implements both synchronous and asynchronous event processing
// Leverages existing database manager, worker pool, and maintains tenant isolation
type HybridEventBus struct {
	// Core infrastructure
	dbManager         *db.Manager
	workerPoolManager *workerpool.WorkerPoolManager
	logger            logger.Logger
	
	// Event handlers
	syncHandlers  map[string][]EventHandler
	asyncHandlers map[string][]EventHandler
	
	// Concurrency control
	mu      sync.RWMutex
	running atomic.Bool
	
	// Optional event storage for audit trail
	eventStore EventStore
	
	// Worker pool for async processing
	eventWorkerPool *workerpool.WorkerPool
}

// EventJob wraps an event handler for worker pool processing
type EventJob struct {
	handler func(context.Context, DomainEvent) error
	event   DomainEvent
	ctx     context.Context
	logger  logger.Logger
}

// Process implements the workerpool.Job interface
func (ej *EventJob) Process(ctx context.Context) error {
	if err := ej.handler(ej.ctx, ej.event); err != nil {
		ej.logger.Error("Error handling event asynchronously",
			logger.Field{Key: "event_type", Value: ej.event.GetEventType()},
			logger.Field{Key: "event_id", Value: ej.event.GetEventID()},
			logger.Field{Key: "error", Value: err.Error()})
		return err
	}
	return nil
}

// NewHybridEventBus creates a new hybrid event bus using existing infrastructure
func NewHybridEventBus(
	dbManager *db.Manager,
	workerPoolManager *workerpool.WorkerPoolManager,
	logger logger.Logger,
) *HybridEventBus {
	bus := &HybridEventBus{
		dbManager:         dbManager,
		workerPoolManager: workerPoolManager,
		logger:            logger,
		syncHandlers:      make(map[string][]EventHandler),
		asyncHandlers:     make(map[string][]EventHandler),
	}

	// Initialize optional event store if database manager is available
	if dbManager != nil {
		bus.eventStore = NewDatabaseEventStore(dbManager, logger)
		logger.Info("Event store initialized successfully")
	} else {
		logger.Info("Event store disabled - no database manager provided")
	}

	return bus
}

// PublishSync publishes events synchronously within the same transaction
// Perfect for critical operations requiring immediate consistency
func (bus *HybridEventBus) PublishSync(ctx context.Context, event DomainEvent) error {
	// Store event for audit trail if event store is available
	if bus.eventStore != nil {
		if err := bus.eventStore.Store(ctx, event); err != nil {
			bus.logger.Warn("Failed to store event for audit",
				logger.Field{Key: "event_type", Value: event.GetEventType()},
				logger.Field{Key: "event_id", Value: event.GetEventID()},
				logger.Field{Key: "error", Value: err.Error()})
		}
	}

	bus.mu.RLock()
	handlers, exists := bus.syncHandlers[event.GetEventType()]
	bus.mu.RUnlock()
	
	if !exists {
		bus.logger.Debug("No synchronous handlers for event type",
			logger.Field{Key: "event_type", Value: event.GetEventType()},
			logger.Field{Key: "tenant_id", Value: event.GetTenantID()})
		return nil
	}
	
	// Process all synchronous handlers in order
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			bus.logger.Error("Synchronous event handler failed",
				logger.Field{Key: "event_type", Value: event.GetEventType()},
				logger.Field{Key: "event_id", Value: event.GetEventID()},
				logger.Field{Key: "error", Value: err.Error()})
			return err // Fail fast for synchronous events
		}
	}
	
	bus.logger.Debug("Synchronous event published successfully",
		logger.Field{Key: "event_type", Value: event.GetEventType()},
		logger.Field{Key: "event_id", Value: event.GetEventID()},
		logger.Field{Key: "handlers_count", Value: len(handlers)})
	
	return nil
}

// PublishAsync publishes events asynchronously using the existing worker pool
// Perfect for non-critical operations that can be eventually consistent
func (bus *HybridEventBus) PublishAsync(ctx context.Context, event DomainEvent) error {
	if !bus.running.Load() {
		return fmt.Errorf("event bus not started")
	}

	// Store event for audit trail if event store is available
	if bus.eventStore != nil {
		if err := bus.eventStore.Store(ctx, event); err != nil {
			bus.logger.Warn("Failed to store event for audit",
				logger.Field{Key: "event_type", Value: event.GetEventType()},
				logger.Field{Key: "event_id", Value: event.GetEventID()},
				logger.Field{Key: "error", Value: err.Error()})
		}
	}

	bus.mu.RLock()
	handlers, exists := bus.asyncHandlers[event.GetEventType()]
	bus.mu.RUnlock()
	
	if !exists {
		bus.logger.Debug("No asynchronous handlers for event type",
			logger.Field{Key: "event_type", Value: event.GetEventType()},
			logger.Field{Key: "tenant_id", Value: event.GetTenantID()})
		return nil
	}
	
	// Submit async handlers to worker pool
	for _, handler := range handlers {
		// Capture handler in closure
		h := handler
		
		// Create a job that wraps the event handler
		job := &EventJob{
			handler: h,
			event:   event,
			ctx:     ctx,
			logger:  bus.logger,
		}
		
		err := bus.eventWorkerPool.Submit(job)
		if err != nil {
			return fmt.Errorf("failed to submit event to worker pool: %w", err)
		}
	}
	
	bus.logger.Debug("Asynchronous event published successfully",
		logger.Field{Key: "event_type", Value: event.GetEventType()},
		logger.Field{Key: "event_id", Value: event.GetEventID()},
		logger.Field{Key: "handlers_count", Value: len(handlers)})
	
	return nil
}

// Subscribe registers a synchronous event handler
func (bus *HybridEventBus) Subscribe(eventType string, handler EventHandler) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	
	bus.syncHandlers[eventType] = append(bus.syncHandlers[eventType], handler)
	
	bus.logger.Info("Synchronous event handler registered",
		logger.Field{Key: "event_type", Value: eventType})
	
	return nil
}

// SubscribeAsync registers an asynchronous event handler
func (bus *HybridEventBus) SubscribeAsync(eventType string, handler EventHandler) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	
	bus.asyncHandlers[eventType] = append(bus.asyncHandlers[eventType], handler)
	
	bus.logger.Info("Asynchronous event handler registered",
		logger.Field{Key: "event_type", Value: eventType})
	
	return nil
}

// Start initializes the event bus and starts worker pools
func (bus *HybridEventBus) Start(ctx context.Context) error {
	if !bus.running.CompareAndSwap(false, true) {
		return nil // Already running
	}
	
	// Initialize worker pool for async event processing
	config := workerpool.WorkerPoolConfig{
		Name:           "events",
		InitialWorkers: 5,
		MaxJobs:        1000,
		MinWorkers:     2,
		MaxWorkers:     10,
		ScaleFactor:    1.5,
		IdleTimeout:    30 * time.Second,
		ScaleInterval:  10 * time.Second,
	}
	
	bus.eventWorkerPool = bus.workerPoolManager.GetOrCreatePool(config)
	
	// Initialize event store if available
	if bus.eventStore != nil {
		if err := bus.eventStore.Initialize(ctx); err != nil {
			bus.logger.Warn("Failed to initialize event store",
				logger.Field{Key: "error", Value: err.Error()})
		} else {
			bus.logger.Info("Event store initialized successfully")
		}
	}
	
	bus.logger.Info("Hybrid event bus started successfully")
	return nil
}

// Stop gracefully shuts down the event bus
func (bus *HybridEventBus) Stop(ctx context.Context) error {
	if !bus.running.CompareAndSwap(true, false) {
		return nil // Already stopped
	}
	
	if bus.eventWorkerPool != nil {
		bus.eventWorkerPool.Stop()
	}
	
	bus.logger.Info("Hybrid event bus stopped successfully")
	return nil
}

// GetEventStore returns the event store if available (optional feature)
func (bus *HybridEventBus) GetEventStore() EventStore {
	return bus.eventStore
}

// SetEventStore allows setting a custom event store (optional)
func (bus *HybridEventBus) SetEventStore(store EventStore) {
	bus.eventStore = store
	if store != nil {
		bus.logger.Info("Custom event store set successfully")
	} else {
		bus.logger.Info("Event store disabled")
	}
}

// EventMetrics provides metrics about event processing (for monitoring)
type EventMetrics struct {
	TotalSyncEvents    int64
	TotalAsyncEvents   int64
	FailedSyncEvents   int64
	FailedAsyncEvents  int64
	AverageProcessTime time.Duration
}

// GetMetrics returns current event bus metrics
func (bus *HybridEventBus) GetMetrics() EventMetrics {
	// Implementation would track metrics over time
	return EventMetrics{}
}
