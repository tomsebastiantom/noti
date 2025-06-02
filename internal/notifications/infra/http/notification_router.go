package notificationroutes

import (
	"net/http"

	"getnoti.com/internal/container"
	sendnotification "getnoti.com/internal/notifications/usecases/send_notification"
	"getnoti.com/internal/shared/handler"
	"getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/shared/utils"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/credentials"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/workerpool"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	BaseHandler       *handler.BaseHandler
	ServiceContainer  *container.ServiceContainer
	GenericCache      *cache.GenericCache
	QueueManager      *queue.QueueManager
	CredentialManager  *credentials.Manager
	WorkerPoolManager *workerpool.WorkerPoolManager
}

func NewHandlers(baseHandler *handler.BaseHandler, serviceContainer *container.ServiceContainer, genericCache *cache.GenericCache, queueManager *queue.QueueManager, credentialManager  *credentials.Manager,wpm *workerpool.WorkerPoolManager) *Handlers {
	return &Handlers{
		BaseHandler:       baseHandler,
		ServiceContainer:  serviceContainer,
		GenericCache:      genericCache,
		QueueManager:      queueManager,
		CredentialManager: credentialManager,
		WorkerPoolManager: wpm,
	}
}

func (h *Handlers) SendNotification(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from request context
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)

	// Get services from container
	providerService := h.ServiceContainer.GetProviderService()
	templateService := h.ServiceContainer.GetTemplateService()

	// Get tenant-specific repositories from container
	notificationRepo, err := h.ServiceContainer.GetNotificationRepositoryForTenant(tenantID)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get notification repository", err, http.StatusInternalServerError)
		return
	}
	
	providerRepo, err := h.ServiceContainer.GetProviderRepositoryForTenant(tenantID)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to get provider repository", err, http.StatusInternalServerError)
		return
	}

	// Initialize use case with container services
	sendNotificationUseCase := sendnotification.NewSendNotificationUseCase(
		providerService, 
		templateService, 
		providerRepo, 
		notificationRepo, 
		h.GenericCache,
	)

	// Initialize controller
	sendNotificationController := sendnotification.NewSendNotificationController(sendNotificationUseCase)

	// Decode the request body
	var req sendnotification.SendNotificationRequest
	if !h.BaseHandler.DecodeJSONBody(w, r, &req) {
		return
	}

	if err := utils.AddTenantIDToRequest(r, &req); err != nil {
		h.BaseHandler.HandleError(w, "Failed to process tenant ID", err, http.StatusInternalServerError)
		return
	}

	// Execute the controller method
	res, err := sendNotificationController.SendNotification(r.Context(), req)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to send notification", err, http.StatusInternalServerError)
		return
	}

	h.BaseHandler.RespondWithJSON(w, res)
}

func NewRouter(serviceContainer *container.ServiceContainer, dbManager *db.Manager, providerCache *cache.GenericCache, queueManager *queue.QueueManager, credentialManager  *credentials.Manager,wpm *workerpool.WorkerPoolManager) *chi.Mux {
	b := handler.NewBaseHandler(dbManager)
	h := NewHandlers(b, serviceContainer, providerCache, queueManager, credentialManager,wpm)

	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.SendNotification)

	// Add more routes here
	// r.Get("/another-route", h.AnotherHandler)

	return r
}
