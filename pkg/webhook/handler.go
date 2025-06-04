package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"getnoti.com/pkg/logger"
)

// WebhookSecretStore interface for retrieving webhook secrets
type WebhookSecretStore interface {
	GetWebhookSecret(ctx context.Context, webhookID string) (string, error)
}

// Handler processes incoming webhooks with security verification
type Handler struct {
	logger       logger.Logger
	secretStore  WebhookSecretStore
	securityMgr  *SecurityManager
	processor    WebhookProcessor
}

// WebhookProcessor interface for processing verified webhooks
type WebhookProcessor interface {
	ProcessWebhook(ctx context.Context, webhook IncomingWebhook) error
}

// NewHandler creates a new webhook handler
func NewHandler(logger logger.Logger, secretStore WebhookSecretStore, processor WebhookProcessor) *Handler {
	return &Handler{
		logger:      logger,
		secretStore: secretStore,
		securityMgr: NewSecurityManager(),
		processor:   processor,
	}
}

// ServeHTTP handles incoming webhook requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Only allow POST requests
	if r.Method != http.MethodPost {
		h.logger.WarnContext(ctx, "Invalid HTTP method for webhook",
			logger.String("method", r.Method),
			logger.String("remote_addr", r.RemoteAddr))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.WarnContext(ctx, "Failed to read webhook body",
			logger.String("error", err.Error()),
			logger.String("remote_addr", r.RemoteAddr))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get signature and timestamp from headers
	signature := r.Header.Get(SignatureHeaderName)
	timestampStr := r.Header.Get(TimestampHeaderName)
	
	if signature == "" || timestampStr == "" {
		h.logger.WarnContext(ctx, "Missing required webhook headers",
			logger.String("remote_addr", r.RemoteAddr),
			logger.Bool("has_signature", signature != ""),
			logger.Bool("has_timestamp", timestampStr != ""))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		h.logger.WarnContext(ctx, "Invalid timestamp in webhook",
			logger.String("timestamp", timestampStr),
			logger.String("remote_addr", r.RemoteAddr))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse webhook payload
	var wh IncomingWebhook
	if err := json.Unmarshal(body, &wh); err != nil {
		h.logger.WarnContext(ctx, "Invalid JSON in webhook payload",
			logger.String("error", err.Error()),
			logger.String("remote_addr", r.RemoteAddr))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate required fields
	if wh.TenantID == "" || wh.EventType == "" || wh.EventID == "" {
		h.logger.WarnContext(ctx, "Missing required fields in webhook",
			logger.String("tenant_id", wh.TenantID),
			logger.String("event_type", wh.EventType),
			logger.String("event_id", wh.EventID),
			logger.String("remote_addr", r.RemoteAddr))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get webhook ID from query parameters or derive from tenant/event
	webhookID := r.URL.Query().Get("webhook_id")
	if webhookID == "" {
		// For backward compatibility, you might derive webhook ID differently
		webhookID = fmt.Sprintf("%s-%s", wh.TenantID, wh.EventType)
	}

	// Get webhook secret for verification
	secret, err := h.secretStore.GetWebhookSecret(ctx, webhookID)
	if err != nil {
		h.logger.WarnContext(ctx, "Failed to get webhook secret",
			logger.String("webhook_id", webhookID),
			logger.String("tenant_id", wh.TenantID),
			logger.String("error", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Verify signature with timestamp for replay protection
	if !h.securityMgr.VerifySignatureWithTimestamp(secret, body, signature, timestamp) {
		h.logger.WarnContext(ctx, "Invalid webhook signature",
			logger.String("webhook_id", webhookID),
			logger.String("tenant_id", wh.TenantID),
			logger.String("remote_addr", r.RemoteAddr))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Set signature in webhook for processor
	wh.Signature = signature

	h.logger.InfoContext(ctx, "Received verified webhook",
		logger.String("webhook_id", webhookID),
		logger.String("tenant_id", wh.TenantID),
		logger.String("event_type", wh.EventType),
		logger.String("event_id", wh.EventID))

	// Process the webhook
	if err := h.processor.ProcessWebhook(ctx, wh); err != nil {
		h.logger.ErrorContext(ctx, "Failed to process webhook",
			logger.String("webhook_id", webhookID),
			logger.String("tenant_id", wh.TenantID),
			logger.String("event_type", wh.EventType),
			logger.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.logger.DebugContext(ctx, "Webhook processed successfully",
		logger.String("webhook_id", webhookID),
		logger.String("tenant_id", wh.TenantID),
		logger.String("event_type", wh.EventType))

	w.WriteHeader(http.StatusOK)
}

// DefaultWebhookProcessor is a simple processor that just logs the webhook
type DefaultWebhookProcessor struct {
	logger logger.Logger
}

// NewDefaultWebhookProcessor creates a default webhook processor
func NewDefaultWebhookProcessor(logger logger.Logger) *DefaultWebhookProcessor {
	return &DefaultWebhookProcessor{logger: logger}
}

// ProcessWebhook processes a webhook by logging it
func (p *DefaultWebhookProcessor) ProcessWebhook(ctx context.Context, webhook IncomingWebhook) error {
	p.logger.InfoContext(ctx, "Processing webhook",
		logger.String("tenant_id", webhook.TenantID),
		logger.String("event_type", webhook.EventType),
		logger.String("event_id", webhook.EventID),
		logger.String("timestamp", webhook.Timestamp.String()))
	return nil
}
