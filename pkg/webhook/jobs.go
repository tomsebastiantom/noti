package webhook

import (
	"context"
	"time"

	"getnoti.com/pkg/logger"
)

// DeliveryJob represents a webhook delivery job for the worker pool
type DeliveryJob struct {
	webhook     OutgoingWebhook
	delivery    *WebhookDelivery
	sender      *Sender
	repository  Repository
	maxAttempts int
	logger      logger.Logger
}

// NewDeliveryJob creates a new webhook delivery job
func NewDeliveryJob(
	webhook OutgoingWebhook,
	delivery *WebhookDelivery,
	sender *Sender,
	repository Repository,
	logger logger.Logger,
) *DeliveryJob {
	return &DeliveryJob{
		webhook:     webhook,
		delivery:    delivery,
		sender:      sender,
		repository:  repository,
		maxAttempts: 3, // Default max attempts
		logger:      logger,
	}
}

// Process implements the workerpool.Job interface
func (j *DeliveryJob) Process(ctx context.Context) error {
	j.logger.DebugContext(ctx, "Processing webhook delivery job",
		logger.String("tenant_id", j.webhook.TenantID),
		logger.String("webhook_id", j.webhook.WebhookID),
		logger.String("delivery_id", j.delivery.ID),
		logger.Int("attempt", j.delivery.AttemptCount))

	// Attempt delivery
	updatedDelivery, err := j.sender.Send(ctx, j.webhook)
	if err != nil {
		j.logger.WarnContext(ctx, "Webhook delivery attempt failed",
			logger.String("tenant_id", j.webhook.TenantID),
			logger.String("webhook_id", j.webhook.WebhookID),
			logger.String("delivery_id", j.delivery.ID),
			logger.String("error", err.Error()),
			logger.Int("attempt", j.delivery.AttemptCount))

		// Update delivery with failure information
		j.delivery.StatusCode = updatedDelivery.StatusCode
		j.delivery.Response = updatedDelivery.Response
		j.delivery.AttemptCount = updatedDelivery.AttemptCount

		// Check if we should schedule a retry
		if j.delivery.AttemptCount < j.maxAttempts && shouldRetry(updatedDelivery.StatusCode) {
			// Calculate next retry time with exponential backoff
			backoffDuration := time.Duration(j.delivery.AttemptCount*j.delivery.AttemptCount) * time.Minute
			nextRetry := time.Now().Add(backoffDuration)
			j.delivery.NextRetryAt = &nextRetry

			j.logger.InfoContext(ctx, "Scheduling webhook delivery retry",
				logger.String("tenant_id", j.webhook.TenantID),
				logger.String("webhook_id", j.webhook.WebhookID),
				logger.String("delivery_id", j.delivery.ID),
				logger.Time("next_retry_at", nextRetry),
				logger.Int("attempt", j.delivery.AttemptCount))
		} else {
			// No more retries, mark as failed
			j.delivery.NextRetryAt = nil
			j.logger.ErrorContext(ctx, "Webhook delivery failed permanently",
				logger.String("tenant_id", j.webhook.TenantID),
				logger.String("webhook_id", j.webhook.WebhookID),
				logger.String("delivery_id", j.delivery.ID),
				logger.Int("final_attempt", j.delivery.AttemptCount))
		}
	} else {
		// Success
		j.delivery.StatusCode = updatedDelivery.StatusCode
		j.delivery.Response = updatedDelivery.Response
		j.delivery.AttemptCount = updatedDelivery.AttemptCount
		j.delivery.DeliveredAt = updatedDelivery.DeliveredAt
		j.delivery.NextRetryAt = nil

		j.logger.InfoContext(ctx, "Webhook delivered successfully",
			logger.String("tenant_id", j.webhook.TenantID),
			logger.String("webhook_id", j.webhook.WebhookID),
			logger.String("delivery_id", j.delivery.ID),
			logger.Int("status_code", j.delivery.StatusCode),
			logger.Int("attempt", j.delivery.AttemptCount))
	}

	// Update delivery record in database
	if updateErr := j.repository.UpdateDelivery(ctx, j.delivery); updateErr != nil {
		j.logger.ErrorContext(ctx, "Failed to update webhook delivery record",
			logger.String("tenant_id", j.webhook.TenantID),
			logger.String("webhook_id", j.webhook.WebhookID),
			logger.String("delivery_id", j.delivery.ID),
			logger.String("error", updateErr.Error()))
		// Return the update error as it's critical for tracking
		return updateErr
	}

	// Return the original delivery error if any (for job retry logic)
	return err
}

// shouldRetry determines if a webhook delivery should be retried based on status code
func shouldRetry(statusCode int) bool {
	// Retry on server errors (5xx) and rate limiting (429)
	// Don't retry on client errors (4xx) except 429
	return statusCode >= 500 || statusCode == 429 || statusCode == 0 // 0 means network error
}

// RetryJob represents a webhook retry job that processes pending retries
type RetryJob struct {
	tenantID   string
	sender     *Sender
	repository Repository
	logger     logger.Logger
}

// NewRetryJob creates a new webhook retry job
func NewRetryJob(
	tenantID string,
	sender *Sender,
	repository Repository,
	logger logger.Logger,
) *RetryJob {
	return &RetryJob{
		tenantID:   tenantID,
		sender:     sender,
		repository: repository,
		logger:     logger,
	}
}

// Process implements the workerpool.Job interface for retry processing
func (j *RetryJob) Process(ctx context.Context) error {
	j.logger.DebugContext(ctx, "Processing webhook retry job",
		logger.String("tenant_id", j.tenantID))

	// Get pending retries for this tenant
	pendingRetries, err := j.repository.GetPendingRetries(ctx, j.tenantID, 50) // Process up to 50 retries
	if err != nil {
		j.logger.ErrorContext(ctx, "Failed to get pending webhook retries",
			logger.String("tenant_id", j.tenantID),
			logger.String("error", err.Error()))
		return err
	}

	if len(pendingRetries) == 0 {
		j.logger.DebugContext(ctx, "No pending webhook retries found",
			logger.String("tenant_id", j.tenantID))
		return nil
	}

	j.logger.InfoContext(ctx, "Processing pending webhook retries",
		logger.String("tenant_id", j.tenantID),
		logger.Int("retry_count", len(pendingRetries)))

	// Process each pending retry
	for _, delivery := range pendingRetries {
		// Get the webhook configuration
		webhook, err := j.repository.GetWebhook(ctx, j.tenantID, delivery.WebhookID)
		if err != nil {
			j.logger.WarnContext(ctx, "Failed to get webhook configuration for retry",
				logger.String("tenant_id", j.tenantID),
				logger.String("webhook_id", delivery.WebhookID),
				logger.String("delivery_id", delivery.ID),
				logger.String("error", err.Error()))
			continue
		}

		// Skip inactive webhooks
		if !webhook.IsActive {
			j.logger.DebugContext(ctx, "Skipping retry for inactive webhook",
				logger.String("tenant_id", j.tenantID),
				logger.String("webhook_id", delivery.WebhookID),
				logger.String("delivery_id", delivery.ID))
			continue
		}

		// Create outgoing webhook from delivery and config
		outgoing := OutgoingWebhook{
			WebhookID: delivery.WebhookID,
			URL:       webhook.URL,
			Payload:   []byte(delivery.Payload),
			Headers:   webhook.Headers,
			TenantID:  delivery.TenantID,
			EventType: delivery.EventType,
			EventID:   delivery.EventID,
			Secret:    webhook.Secret,
		}

		// Create and process delivery job
		deliveryJob := NewDeliveryJob(outgoing, delivery, j.sender, j.repository, j.logger)
		if err := deliveryJob.Process(ctx); err != nil {
			j.logger.WarnContext(ctx, "Webhook retry failed",
				logger.String("tenant_id", j.tenantID),
				logger.String("webhook_id", delivery.WebhookID),
				logger.String("delivery_id", delivery.ID),
				logger.String("error", err.Error()))
		}
	}

	j.logger.InfoContext(ctx, "Completed processing webhook retries",
		logger.String("tenant_id", j.tenantID),
		logger.Int("processed_count", len(pendingRetries)))

	return nil
}
