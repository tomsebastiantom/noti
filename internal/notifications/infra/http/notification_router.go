package notificationroutes

import (
	"net/http"

	"getnoti.com/internal/notifications/repos/implementations"
	"getnoti.com/internal/notifications/usecases/send_notification"
	"getnoti.com/internal/providers/infra/providers"
	providerrepos "getnoti.com/internal/providers/repos/implementations"
	providerService "getnoti.com/internal/providers/services"
	"getnoti.com/internal/shared/handler"
	"getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/shared/utils"
	templatesrepo "getnoti.com/internal/templates/repos/implementations"
	templates "getnoti.com/internal/templates/services"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/credentials"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/workerpool"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	BaseHandler       *handler.BaseHandler
	GenericCache      *cache.GenericCache
	QueueManager      *queue.QueueManager
	CredentialManager  *credentials.Manager
	WorkerPoolManager *workerpool.WorkerPoolManager
}

func NewHandlers(baseHandler *handler.BaseHandler, genericCache *cache.GenericCache, queueManager *queue.QueueManager, credentialManager  *credentials.Manager,wpm *workerpool.WorkerPoolManager) *Handlers {
	return &Handlers{
		BaseHandler:       baseHandler,
		GenericCache:      genericCache,
		QueueManager:      queueManager,
		CredentialManager: credentialManager,
		WorkerPoolManager: wpm,
	}
}

func (h *Handlers) SendNotification(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.BaseHandler.DBManager.GetDatabaseConnection(tenantID)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve database connection", err, http.StatusInternalServerError)
		return
	}

	notificationQueue, err := h.QueueManager.GetOrCreateQueue(tenantID)
	if err != nil {
		h.BaseHandler.HandleError(w, "Failed to retrieve or create notification queue", err, http.StatusInternalServerError)
		return
	}

	// Initialize repositories
	notificationRepo := repos.NewNotificationRepository(database)
	templatesRepo := templatesrepo.NewTemplateRepository(database)
	providerRepo := providerrepos.NewProviderRepository(database)

	// Initialize services
	providerFactory := providers.NewProviderFactory(h.GenericCache, providerRepo,h.CredentialManager)
	providerService := providerService.NewProviderService(providerFactory, notificationQueue, h.WorkerPoolManager)
	templateService := templates.NewTemplateService(templatesRepo)

	// Initialize use case
	sendNotificationUseCase := sendnotification.NewSendNotificationUseCase(providerService, templateService, providerRepo, notificationRepo, h.GenericCache)

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

func NewRouter(dbManager *db.Manager, providerCache *cache.GenericCache, queueManager *queue.QueueManager, credentialManager  *credentials.Manager,wpm *workerpool.WorkerPoolManager) *chi.Mux {
	b := handler.NewBaseHandler(dbManager)
	h := NewHandlers(b, providerCache, queueManager, credentialManager,wpm)

	r := chi.NewRouter()

	// Set up routes
	r.Post("/", h.SendNotification)

	// Add more routes here
	// r.Get("/another-route", h.AnotherHandler)

	return r
}
