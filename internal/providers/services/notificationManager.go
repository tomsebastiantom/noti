package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"getnoti.com/internal/providers/dtos"
	"getnoti.com/internal/providers/infra/providers"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/workerpool"
)

type NotificationManager struct {
	notificationQueue queue.Queue
	providerFactory   *providers.ProviderFactory
	logger            logger.Logger
	workerPools       map[string]*workerpool.WorkerPool
	mu                sync.RWMutex
}

func NewNotificationManager(nq queue.Queue, pf *providers.ProviderFactory, log logger.Logger) *NotificationManager {
	return &NotificationManager{
		notificationQueue: nq,
		providerFactory:   pf,
		logger:            log,
		workerPools:       make(map[string]*workerpool.WorkerPool),
	}
}

func (nm *NotificationManager) SendNotification(ctx context.Context, req dtos.SendNotificationRequest) error {
	messageBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %w", err)
	}

	err = nm.notificationQueue.Publish(ctx, req.ProviderID, req.ProviderID, queue.Message{Body: messageBytes})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	nm.ensureConsumerAndWorkerPool(ctx, req.ProviderID)
	return nil
}

func (nm *NotificationManager) ensureConsumerAndWorkerPool(ctx context.Context, providerID string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, exists := nm.workerPools[providerID]; !exists {
		wp := workerpool.NewWorkerPool(5, 100, 1, 50, 0.1, nm.logger)
		nm.workerPools[providerID] = wp

		err := nm.notificationQueue.DeclareQueue(ctx, providerID, providerID, true, false, false)
		if err != nil {
			nm.logger.Error("Failed to declare queue for provider %s: %v", providerID, err)
			return
		}

		err = nm.notificationQueue.InitializeConsumer(ctx, providerID, providerID, func(msg queue.Message) {
			nm.handleMessage(providerID, msg)
		})
		if err != nil {
			nm.logger.Error("Failed to initialize consumer for provider %s: %v", providerID, err)
			return
		}

		nm.logger.Info("Created consumer and worker pool for provider %s", providerID)
	}
}

func (nm *NotificationManager) handleMessage(providerID string, msg queue.Message) {
	var req dtos.SendNotificationRequest
	err := json.Unmarshal(msg.Body, &req)
	if err != nil {
		nm.logger.Error("Failed to unmarshal notification request: %v", err)
		return
	}

	job := &NotificationJob{
		req:             req,
		providerFactory: nm.providerFactory,
		logger:          nm.logger,
	}

	nm.mu.RLock()
	wp, exists := nm.workerPools[providerID]
	nm.mu.RUnlock()

	if exists {
		wp.Submit(job)
	} else {
		nm.logger.Error("Worker pool not found for provider %s", providerID)
	}
}

type NotificationJob struct {
	req             dtos.SendNotificationRequest
	providerFactory *providers.ProviderFactory
	logger          logger.Logger
}

func (j *NotificationJob) Process() {
	provider := j.providerFactory.GetProvider(j.req.ProviderID, j.req.Sender, j.req.Channel)
	if provider == nil {
		j.logger.Error("Failed to get provider instance for provider %s", j.req.ProviderID)
		return
	}

	resp := provider.SendNotification(context.Background(), j.req)
	if !resp.Success {
		j.logger.Error("Failed to send notification: %s", resp.Message)
		// Implement retry logic here if needed
	}
}

func (nm *NotificationManager) Shutdown() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	for providerID, wp := range nm.workerPools {
		wp.Stop()
		nm.logger.Info("Shutting down worker pool for provider %s", providerID)
	}
}

// This implementation:
// Creates a consumer for each provider ID (channel).
// Uses a worker pool for each provider to process messages.
// Ensures that a consumer and worker pool are created for each provider when sending a notification.
// Uses the existing queue methods for publishing, declaring queues, and initializing consumers.
// Key points:
// The ensureConsumerAndWorkerPool function creates a consumer and worker pool for each provider if they don't exist.
// The handleMessage function is used as the consumer handler and submits jobs to the appropriate worker pool.
// The NotificationJob struct represents a job that will be processed by the worker pool.
// The Shutdown method ensures proper cleanup of worker pools.
