package webhookroutes

import (
	"net/http"

	"getnoti.com/internal/container"
	"getnoti.com/internal/shared/handler"
	"getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/webhooks/domain"
	createwebhook "getnoti.com/internal/webhooks/usecases/create_webhook"
	"getnoti.com/pkg/db"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	BaseHandler      *handler.BaseHandler
	ServiceContainer *container.ServiceContainer
}

func NewHandlers(baseHandler *handler.BaseHandler, serviceContainer *container.ServiceContainer) *Handlers {
	return &Handlers{
		BaseHandler:      baseHandler,
		ServiceContainer: serviceContainer,
	}
}

// CreateWebhook handles HTTP requests to create webhooks
func (h *Handlers) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	// Get webhook service from container
	webhookService := h.ServiceContainer.GetWebhookService()

	// Get logger from infrastructure
	infrastructure, err := h.ServiceContainer.GetInfrastructure()
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get infrastructure", err, http.StatusInternalServerError)
		return
	}
	
	// Initialize use case with proper logger
	createWebhookUseCase := createwebhook.NewUseCase(webhookService, infrastructure.Logger)
	createWebhookController := createwebhook.NewController(createWebhookUseCase, infrastructure.Logger, h.BaseHandler)

	// Process the request
	createWebhookController.Handle(w, r)
}

// GetWebhooks handles HTTP requests to list webhooks for a tenant
func (h *Handlers) GetWebhooks(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from request context
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)

	// Get webhook service from container
	webhookService := h.ServiceContainer.GetWebhookService()

	// Get webhooks for tenant using service
	webhooks, _, err := webhookService.ListWebhooks(r.Context(), tenantID, 100, 0) // Default pagination
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get webhooks", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, map[string]interface{}{
		"webhooks": webhooks,
	})
}

// GetWebhook handles HTTP requests to get a specific webhook
func (h *Handlers) GetWebhook(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID and webhook ID from request
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)
	webhookID := chi.URLParam(r, "id")

	if webhookID == "" {
		h.BaseHandler.HandleError(w, "Webhook ID is required", nil, http.StatusBadRequest)
		return
	}

	// Get webhook service from container
	webhookService := h.ServiceContainer.GetWebhookService()

	// Get webhook by ID using service
	webhook, err := webhookService.GetWebhook(r.Context(), tenantID, webhookID)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get webhook", err, http.StatusNotFound)
		return
	}

	h.BaseHandler.RespondWithJSON(w, webhook)
}

// UpdateWebhook handles HTTP requests to update webhooks
func (h *Handlers) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID and webhook ID from request
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)
	webhookID := chi.URLParam(r, "id")

	if webhookID == "" {
		h.BaseHandler.HandleError(w, "Webhook ID is required", nil, http.StatusBadRequest)
		return
	}

	// Parse request body
	var updateReq struct {
		URL       string   `json:"url"`
		Events    []string `json:"events"`
		IsActive  bool     `json:"is_active"`
		Secret    string   `json:"secret,omitempty"`
	}
	if !h.BaseHandler.DecodeJSONBody(w, r, &updateReq) {
		return
	}

	// Get webhook service from container
	webhookService := h.ServiceContainer.GetWebhookService()

	// Create updates domain object
	updates := &domain.Webhook{
		URL:      updateReq.URL,
		Events:   updateReq.Events,
		IsActive: updateReq.IsActive,
		Secret:   updateReq.Secret,
	}

	// Update webhook using service
	updatedWebhook, err := webhookService.UpdateWebhook(r.Context(), tenantID, webhookID, updates)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to update webhook", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, updatedWebhook)
}

// DeleteWebhook handles HTTP requests to delete webhooks
func (h *Handlers) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID and webhook ID from request
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)
	webhookID := chi.URLParam(r, "id")

	if webhookID == "" {
		h.BaseHandler.HandleError(w, "Webhook ID is required", nil, http.StatusBadRequest)
		return
	}

	// Get webhook service from container
	webhookService := h.ServiceContainer.GetWebhookService()

	// Delete webhook using service
	err := webhookService.DeleteWebhook(r.Context(), tenantID, webhookID)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to delete webhook", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, map[string]string{
		"message": "Webhook deleted successfully",
	})
}

// NewRouter sets up the router with all webhook routes
func NewRouter(serviceContainer *container.ServiceContainer, dbManager *db.Manager) *chi.Mux {
	b := handler.NewBaseHandler(dbManager)
	h := NewHandlers(b, serviceContainer)

	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.CreateWebhook)
	r.Get("/", h.GetWebhooks)
	r.Get("/{id}", h.GetWebhook)
	r.Put("/{id}", h.UpdateWebhook)
	r.Delete("/{id}", h.DeleteWebhook)

	return r
}
