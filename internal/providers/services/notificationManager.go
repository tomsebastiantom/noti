package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"getnoti.com/internal/providers/dtos"
	"getnoti.com/internal/providers/infra/providers"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/workerpool"
)

type NotificationManager struct {
	notificationQueue queue.Queue
	providerFactory   *providers.ProviderFactory
	workerPoolManager *workerpool.WorkerPoolManager
	mu                sync.RWMutex
}

func NewNotificationManager(nq queue.Queue, pf *providers.ProviderFactory, wpm *workerpool.WorkerPoolManager) *NotificationManager {
	return &NotificationManager{
		notificationQueue: nq,
		providerFactory:   pf,
		workerPoolManager: wpm,
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

	return nm.ensureConsumerAndWorkerPool(ctx, req.TenantID, req.ProviderID)
}

func (nm *NotificationManager) ensureConsumerAndWorkerPool(ctx context.Context, tenantID string, providerID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	pool, exists := nm.workerPoolManager.GetPool(providerID)
	if !exists {
		config := workerpool.WorkerPoolConfig{
			Name:           providerID,
			InitialWorkers: 5,
			MaxJobs:        100,
			MinWorkers:     1,
			MaxWorkers:     50,
			ScaleFactor:    0.1,
			IdleTimeout:    5 * time.Minute,
			ScaleInterval:  30 * time.Second,
		}
		pool = nm.workerPoolManager.GetOrCreatePool(config)

		err := nm.notificationQueue.DeclareQueue(ctx, providerID, tenantID, true, false, false)
		if err != nil {
			return fmt.Errorf("failed to declare queue: %w", err)
		}

		err = nm.notificationQueue.InitializeConsumer(ctx, providerID, tenantID, func(msg queue.Message) {
			nm.handleMessage(providerID, msg)
		}, pool)
		if err != nil {
			return fmt.Errorf("failed to initialize consumer: %w", err)
		}
	}
	return nil
}

func (nm *NotificationManager) handleMessage(providerID string, msg queue.Message) {
	var req dtos.SendNotificationRequest
	err := json.Unmarshal(msg.Body, &req)
	if err != nil {
		fmt.Printf("Failed to unmarshal message: %v, providerID: %s\n", err, providerID)
		return
	}

	job := &NotificationJob{
		req:             req,
		providerFactory: nm.providerFactory,
	}

	pool, exists := nm.workerPoolManager.GetPool(providerID)
	if !exists {
		fmt.Printf("Worker pool not found for providerID: %s\n", providerID)
		return
	}

	err = pool.Submit(job)
	if err != nil {
		fmt.Printf("Failed to submit job to worker pool: %v, providerID: %s\n", err, providerID)
	}
}

type NotificationJob struct {
	req             dtos.SendNotificationRequest
	providerFactory *providers.ProviderFactory
}

func (j *NotificationJob) Process(ctx context.Context) error {
	provider := j.providerFactory.GetProvider(j.req.ProviderID, j.req.Sender, j.req.Channel)
	if provider == nil {
		return fmt.Errorf("failed to get provider instance for provider %s", j.req.ProviderID)
	}

	resp := provider.SendNotification(ctx, j.req)
	if !resp.Success {
		return fmt.Errorf("failed to send notification: %s", resp.Message)
	}
	return nil
}

func (nm *NotificationManager) Shutdown() {
	nm.workerPoolManager.ShutdownAll()
}
