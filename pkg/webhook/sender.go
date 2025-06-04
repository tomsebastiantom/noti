package webhook

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"getnoti.com/pkg/circuitbreaker"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/errors"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/workerpool"
)

// Sender handles webhook delivery with proper error handling, circuit breakers, and worker pool integration
type Sender struct {
	client          *http.Client
	securityManager *SecurityManager
	circuitBreakers map[string]*circuitbreaker.CircuitBreaker // per-tenant circuit breakers
	cbMutex         sync.RWMutex                              // protects circuitBreakers map
	dbManager       *db.Manager
	workerPoolMgr   *workerpool.WorkerPoolManager
	logger          logger.Logger
}

// NewSender creates a new webhook sender with proper dependencies
func NewSender(
	dbManager *db.Manager,
	workerPoolMgr *workerpool.WorkerPoolManager,
	logger logger.Logger,
) *Sender {
	return &Sender{
		client:          &http.Client{Timeout: 30 * time.Second},
		securityManager: NewSecurityManager(),
		circuitBreakers: make(map[string]*circuitbreaker.CircuitBreaker),
		dbManager:       dbManager,
		workerPoolMgr:   workerPoolMgr,
		logger:          logger,
	}
}

// getCircuitBreaker gets or creates a circuit breaker for a tenant using the proper circuit breaker implementation
func (s *Sender) getCircuitBreaker(tenantID string) *circuitbreaker.CircuitBreaker {
	s.cbMutex.RLock()
	cb, ok := s.circuitBreakers[tenantID]
	s.cbMutex.RUnlock()
	
	if !ok {
		s.cbMutex.Lock()
		// Double-check after acquiring write lock
		if cb, ok = s.circuitBreakers[tenantID]; !ok {
			// Create new circuit breaker with appropriate thresholds
			cb = circuitbreaker.NewCircuitBreaker(
				5,               // failure threshold
				3,               // success threshold  
				5*time.Minute,   // reset timeout
			)
			s.circuitBreakers[tenantID] = cb
		}
		s.cbMutex.Unlock()
	}
	return cb
}

// Send delivers a webhook with proper error handling, circuit breaker protection, and structured logging
func (s *Sender) Send(ctx context.Context, wh OutgoingWebhook) (*WebhookDelivery, error) {
	// Create delivery record with unique ID
	deliveryID, err := s.generateDeliveryID()
	if err != nil {
		return nil, errors.New(errors.ErrCodeInternal).
			WithContext(ctx).
			WithOperation("Send").
			WithCause(err).
			WithMessage("Failed to generate delivery ID").
			WithDetails(map[string]interface{}{
				"tenant_id":  wh.TenantID,
				"webhook_id": wh.WebhookID,
			}).
			Build()
	}

	delivery := &WebhookDelivery{
		ID:           deliveryID,
		WebhookID:    wh.WebhookID,
		TenantID:     wh.TenantID,
		EventType:    wh.EventType,
		EventID:      wh.EventID,
		Payload:      string(wh.Payload),
		AttemptCount: 0,
		CreatedAt:    time.Now(),
	}

	cb := s.getCircuitBreaker(wh.TenantID)
	
	// Use circuit breaker to protect against repeated failures
	err = cb.Execute(func() error {
		return s.performDelivery(ctx, wh, delivery)
	})

	if err != nil {
		// Check if this is a circuit breaker error
		if err.Error() == "circuit breaker is open" {
			s.logger.WarnContext(ctx, "Circuit breaker open for tenant webhook delivery",
				logger.String("tenant_id", wh.TenantID),
				logger.String("webhook_id", wh.WebhookID),
				logger.String("delivery_id", delivery.ID))
			
			delivery.StatusCode = 503
			delivery.Response = "circuit breaker open"
			
			return delivery, errors.New(errors.ErrCodeCircuitBreaker).
				WithContext(ctx).
				WithOperation("Send").
				WithMessage("Circuit breaker open for tenant webhook delivery").
				WithDetails(map[string]interface{}{
					"tenant_id":   wh.TenantID,
					"webhook_id":  wh.WebhookID,
					"delivery_id": delivery.ID,
				}).
				Build()
		}
		
		// Other delivery errors are already structured, return as-is
		return delivery, err
	}

	return delivery, nil
}
// performDelivery handles the actual HTTP delivery with retries
func (s *Sender) performDelivery(ctx context.Context, wh OutgoingWebhook, delivery *WebhookDelivery) error {
	// Create timestamp for replay protection
	timestamp := time.Now().Unix()
	
	// Sign payload with timestamp
	signature := s.securityManager.SignPayloadWithTimestamp(wh.Secret, wh.Payload, timestamp)

	// Prepare request
	req, err := http.NewRequestWithContext(ctx, "POST", wh.URL, bytes.NewReader(wh.Payload))
	if err != nil {
		delivery.StatusCode = 0
		delivery.Response = fmt.Sprintf("failed to create request: %v", err)
		
		return errors.New(errors.ErrCodeHTTP).
			WithContext(ctx).
			WithOperation("performDelivery").
			WithCause(err).
			WithMessage("Failed to create HTTP request").
			WithDetails(map[string]interface{}{
				"tenant_id":   wh.TenantID,
				"webhook_id":  wh.WebhookID,
				"delivery_id": delivery.ID,
				"url":         wh.URL,
			}).
			Build()
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "noti-webhook/1.0")
	req.Header.Set(SignatureHeaderName, signature)
	req.Header.Set(TimestampHeaderName, strconv.FormatInt(timestamp, 10))

	// Set custom headers
	for k, v := range wh.Headers {
		req.Header.Set(k, v)
	}

	// Attempt delivery with retries
	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		delivery.AttemptCount = attempt

		s.logger.DebugContext(ctx, "Attempting webhook delivery",
			logger.String("tenant_id", wh.TenantID),
			logger.String("webhook_id", wh.WebhookID),
			logger.String("delivery_id", delivery.ID),
			logger.String("url", wh.URL),
			logger.Int("attempt", attempt))

		resp, err := s.client.Do(req)
		if err != nil {
			delivery.Response = fmt.Sprintf("request failed: %v", err)
			s.logger.WarnContext(ctx, "Webhook delivery request failed",
				logger.String("tenant_id", wh.TenantID),
				logger.String("webhook_id", wh.WebhookID),
				logger.String("delivery_id", delivery.ID),
				logger.String("error", err.Error()),
				logger.Int("attempt", attempt))

			if attempt < maxAttempts {
				backoff := time.Duration(attempt*attempt) * time.Second
				select {
				case <-ctx.Done():
					return errors.New(errors.ErrCodeTimeout).
						WithContext(ctx).
						WithOperation("performDelivery").
						WithCause(ctx.Err()).
						WithMessage("Context cancelled during webhook delivery").
						WithDetails(map[string]interface{}{
							"tenant_id":   wh.TenantID,
							"webhook_id":  wh.WebhookID,
							"delivery_id": delivery.ID,
							"attempt":     attempt,
						}).
						Build()
				case <-time.After(backoff):
					continue
				}
			}
			continue
		}

		// Read response body
		responseBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		delivery.StatusCode = resp.StatusCode
		delivery.Response = string(responseBody)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success
			successTime := time.Now()
			delivery.DeliveredAt = &successTime
			
			s.logger.InfoContext(ctx, "Webhook delivered successfully",
				logger.String("tenant_id", wh.TenantID),
				logger.String("webhook_id", wh.WebhookID),
				logger.String("delivery_id", delivery.ID),
				logger.String("url", wh.URL),
				logger.Int("status_code", resp.StatusCode),
				logger.Int("attempt", attempt))
			
			return nil
		}

		s.logger.WarnContext(ctx, "Webhook delivery failed",
			logger.String("tenant_id", wh.TenantID),
			logger.String("webhook_id", wh.WebhookID),
			logger.String("delivery_id", delivery.ID),
			logger.String("url", wh.URL),
			logger.Int("status_code", resp.StatusCode),
			logger.String("response", string(responseBody)),
			logger.Int("attempt", attempt))

		// Check if we should retry based on status code
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			// Client error (except rate limit), don't retry
			break
		}

		if attempt < maxAttempts {
			backoff := time.Duration(attempt*attempt) * time.Second
			nextRetry := time.Now().Add(backoff)
			delivery.NextRetryAt = &nextRetry
			
			select {
			case <-ctx.Done():
				return errors.New(errors.ErrCodeTimeout).
					WithContext(ctx).
					WithOperation("performDelivery").
					WithCause(ctx.Err()).
					WithMessage("Context cancelled during webhook delivery").
					WithDetails(map[string]interface{}{
						"tenant_id":   wh.TenantID,
						"webhook_id":  wh.WebhookID,
						"delivery_id": delivery.ID,
						"attempt":     attempt,
					}).
					Build()
			case <-time.After(backoff):
				continue
			}
		}
	}

	// All attempts failed
	s.logger.ErrorContext(ctx, "Webhook delivery failed after all attempts",
		logger.String("tenant_id", wh.TenantID),
		logger.String("webhook_id", wh.WebhookID),
		logger.String("delivery_id", delivery.ID),
		logger.String("url", wh.URL),
		logger.Int("attempts", maxAttempts),
		logger.Int("final_status_code", delivery.StatusCode))

	return errors.New(errors.ErrCodeHTTP).
		WithContext(ctx).
		WithOperation("performDelivery").
		WithMessage("Webhook delivery failed after all attempts").
		WithDetails(map[string]interface{}{
			"tenant_id":     wh.TenantID,
			"webhook_id":    wh.WebhookID,
			"delivery_id":   delivery.ID,
			"url":           wh.URL,
			"attempts":      maxAttempts,
			"status_code":   delivery.StatusCode,
			"response":      delivery.Response,		}).
		Build()
}

// SendEvent creates and sends a webhook for an event
func (s *Sender) SendEvent(ctx context.Context, webhookConfig WebhookConfig, event WebhookEvent) (*WebhookDelivery, error) {
	s.logger.DebugContext(ctx, "Sending webhook event",
		logger.String("tenant_id", webhookConfig.TenantID),
		logger.String("webhook_id", webhookConfig.ID),
		logger.String("event_type", event.EventType),
		logger.String("event_id", event.EventID))

	// Create payload
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, errors.New(errors.ErrCodeValidation).
			WithContext(ctx).
			WithOperation("SendEvent").
			WithCause(err).
			WithMessage("Failed to marshal event payload").
			WithDetails(map[string]interface{}{
				"tenant_id":  webhookConfig.TenantID,
				"webhook_id": webhookConfig.ID,
				"event_type": event.EventType,
				"event_id":   event.EventID,
			}).
			Build()
	}

	// Create outgoing webhook
	outgoing := OutgoingWebhook{
		WebhookID: webhookConfig.ID,
		URL:       webhookConfig.URL,
		Payload:   payload,
		Headers:   webhookConfig.Headers,
		TenantID:  webhookConfig.TenantID,
		EventType: event.EventType,
		EventID:   event.EventID,
		Secret:    webhookConfig.Secret,
	}

	return s.Send(ctx, outgoing)
}

// SendEventAsync sends a webhook event asynchronously using the worker pool
func (s *Sender) SendEventAsync(ctx context.Context, webhookConfig WebhookConfig, event WebhookEvent) error {
	s.logger.DebugContext(ctx, "Sending webhook event asynchronously",
		logger.String("tenant_id", webhookConfig.TenantID),
		logger.String("webhook_id", webhookConfig.ID),
		logger.String("event_type", event.EventType),
		logger.String("event_id", event.EventID))

	// Create payload
	payload, err := json.Marshal(event)
	if err != nil {
		return errors.New(errors.ErrCodeValidation).
			WithContext(ctx).
			WithOperation("SendEventAsync").
			WithCause(err).
			WithMessage("Failed to marshal event payload").
			WithDetails(map[string]interface{}{
				"tenant_id":  webhookConfig.TenantID,
				"webhook_id": webhookConfig.ID,
				"event_type": event.EventType,
				"event_id":   event.EventID,
			}).
			Build()
	}

	// Create outgoing webhook
	outgoing := OutgoingWebhook{
		WebhookID: webhookConfig.ID,
		URL:       webhookConfig.URL,
		Payload:   payload,
		Headers:   webhookConfig.Headers,
		TenantID:  webhookConfig.TenantID,
		EventType: event.EventType,
		EventID:   event.EventID,
		Secret:    webhookConfig.Secret,
	}

	// Create delivery record
	deliveryID, err := s.generateDeliveryID()
	if err != nil {
		return errors.New(errors.ErrCodeInternal).
			WithContext(ctx).
			WithOperation("SendEventAsync").
			WithCause(err).
			WithMessage("Failed to generate delivery ID").
			WithDetails(map[string]interface{}{
				"tenant_id":  webhookConfig.TenantID,
				"webhook_id": webhookConfig.ID,
			}).
			Build()
	}

	delivery := &WebhookDelivery{
		ID:           deliveryID,
		WebhookID:    webhookConfig.ID,
		TenantID:     webhookConfig.TenantID,
		EventType:    event.EventType,
		EventID:      event.EventID,
		Payload:      string(payload),
		AttemptCount: 0,
		CreatedAt:    time.Now(),
	}

	// Get tenant-specific database connection
	db, err := s.dbManager.GetDatabaseConnection(webhookConfig.TenantID)
	if err != nil {
		return errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("SendEventAsync").
			WithCause(err).
			WithMessage("Failed to get database connection").
			WithDetails(map[string]interface{}{
				"tenant_id": webhookConfig.TenantID,
			}).
			Build()
	}

	// Create repository
	repository := NewSQLRepository(db, s.logger)

	// Create delivery job
	job := NewDeliveryJob(outgoing, delivery, s, repository, s.logger)
	// Get or create worker pool for this tenant
	poolName := fmt.Sprintf("webhook-%s", webhookConfig.TenantID)
	pool := s.workerPoolMgr.GetOrCreatePool(workerpool.WorkerPoolConfig{
		Name:           poolName,
		InitialWorkers: 2,
		MaxJobs:        100,
		MinWorkers:     1,
		MaxWorkers:     10,
		ScaleFactor:    0.8,
		IdleTimeout:    5 * time.Minute,
		ScaleInterval:  30 * time.Second,
	})

	// Submit job to worker pool
	if err := pool.Submit(job); err != nil {
		return errors.New(errors.ErrCodeQueue).
			WithContext(ctx).
			WithOperation("SendEventAsync").
			WithCause(err).
			WithMessage("Failed to submit webhook job to worker pool").
			WithDetails(map[string]interface{}{
				"tenant_id":   webhookConfig.TenantID,
				"webhook_id":  webhookConfig.ID,
				"delivery_id": delivery.ID,
				"pool_name":   poolName,
			}).
			Build()
	}

	s.logger.InfoContext(ctx, "Webhook event submitted for async delivery",
		logger.String("tenant_id", webhookConfig.TenantID),
		logger.String("webhook_id", webhookConfig.ID),
		logger.String("delivery_id", delivery.ID),
		logger.String("event_type", event.EventType),
		logger.String("event_id", event.EventID))

	return nil
}

// generateDeliveryID generates a unique delivery ID
func (s *Sender) generateDeliveryID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
