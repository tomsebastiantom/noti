package create_webhook

import (
	"net/http"

	"getnoti.com/internal/shared/handler"
	"getnoti.com/internal/shared/middleware"
	"getnoti.com/pkg/logger"
)

// Controller handles HTTP requests for webhook creation
type Controller struct {
	useCase     *UseCase
	logger      logger.Logger
	baseHandler *handler.BaseHandler
}

// NewController creates a new create webhook controller
func NewController(useCase *UseCase, logger logger.Logger, baseHandler *handler.BaseHandler) *Controller {
	return &Controller{
		useCase:     useCase,
		logger:      logger,
		baseHandler: baseHandler,
	}
}

// Handle processes HTTP requests to create webhooks
func (c *Controller) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get tenant ID from context (set by middleware)
	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(string)
	if !ok {
		c.logger.Warn("Tenant ID not found in context")
		c.baseHandler.HandleError(w, "Tenant ID is required", nil, http.StatusBadRequest)
		return
	}

	// Parse request body
	var req CreateWebhookRequest
	if !c.baseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	// Execute use case
	response, err := c.useCase.Execute(ctx, tenantID, &req)
	if err != nil {
		c.handleError(w, tenantID, err)
		return
	}
	// Respond with success
	c.baseHandler.RespondWithJSON(w, response)
}

// handleError handles different types of errors appropriately
func (c *Controller) handleError(w http.ResponseWriter, tenantID string, err error) {
	if webhookErr, ok := err.(*CreateWebhookError); ok {
		switch webhookErr.Code {
		case "VALIDATION_ERROR":
			c.baseHandler.HandleError(w, webhookErr.Message, err, http.StatusBadRequest)
		case "DUPLICATE_NAME":
			c.baseHandler.HandleError(w, webhookErr.Message, err, http.StatusConflict)
		case "UNAUTHORIZED":
			c.baseHandler.HandleError(w, webhookErr.Message, err, http.StatusUnauthorized)
		default:
			c.logger.Error("Webhook creation error", logger.String("tenant_id", tenantID), logger.Err(err))
			c.baseHandler.HandleError(w, "Internal server error", err, http.StatusInternalServerError)
		}
	} else {
		c.logger.Error("Unexpected error during webhook creation", logger.String("tenant_id", tenantID), logger.Err(err))
		c.baseHandler.HandleError(w, "Internal server error", err, http.StatusInternalServerError)
	}
}
