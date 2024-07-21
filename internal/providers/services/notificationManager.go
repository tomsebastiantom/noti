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

    nm.ensureConsumerAndWorkerPool(ctx, req.ProviderID)
    return nil
}

func (nm *NotificationManager) ensureConsumerAndWorkerPool(ctx context.Context, providerID string) {
    nm.mu.Lock()
    defer nm.mu.Unlock()

    _, exists := nm.workerPoolManager.GetPool(providerID)
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
        nm.workerPoolManager.GetOrCreatePool(config)

        err := nm.notificationQueue.DeclareQueue(ctx, providerID, providerID, true, false, false)
        if err != nil {
            // Handle error (consider using a logger or error channel)
            return
        }

        err = nm.notificationQueue.InitializeConsumer(ctx, providerID, providerID, func(msg queue.Message) {
            nm.handleMessage(providerID, msg)
        })
        if err != nil {
            // Handle error (consider using a logger or error channel)
            return
        }
    }
}

func (nm *NotificationManager) handleMessage(providerID string, msg queue.Message) {
    var req dtos.SendNotificationRequest
    err := json.Unmarshal(msg.Body, &req)
    if err != nil {
        // Handle error (consider using a logger or error channel)
        return
    }

    job := &NotificationJob{
        req:             req,
        providerFactory: nm.providerFactory,
    }

    pool, exists := nm.workerPoolManager.GetPool(providerID)
    if exists {
        err := pool.Submit(job)
        if err != nil {
            // Handle error (consider using a logger or error channel)
        }
    } else {
        // Handle error: worker pool not found (consider using a logger or error channel)
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
