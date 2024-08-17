package notificationroutes

import (
	"encoding/json"
	"net/http"

	"getnoti.com/internal/notifications/repos/implementations"
	"getnoti.com/internal/notifications/usecases/send_notification"
	"getnoti.com/internal/providers/infra/providers"
	providerrepos "getnoti.com/internal/providers/repos/implementations"
	providerService "getnoti.com/internal/providers/services"
	"getnoti.com/internal/shared/middleware"
	templatesrepo "getnoti.com/internal/templates/repos/implementations"
	templates "getnoti.com/internal/templates/services"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/workerpool"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	DBManager         *db.Manager
	GenericCache      *cache.GenericCache
	queueManager      *queue.QueueManager
	workerPoolManager *workerpool.WorkerPoolManager
}

func NewHandlers(dbManager *db.Manager, genericCache *cache.GenericCache, queueManager *queue.QueueManager, wpm *workerpool.WorkerPoolManager) *Handlers {
	return &Handlers{
		DBManager:         dbManager,
		GenericCache:      genericCache,
		queueManager:      queueManager,
		workerPoolManager: wpm,
	}
}

func (h *Handlers) SendNotification(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value(middleware.TenantIDKey).(string)

	// Retrieve the database connection
	database, err := h.DBManager.GetDatabaseConnection(tenantID)
	if err != nil {
		http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
		return
	}

	notificationQueue, err := h.queueManager.GetOrCreateQueue(tenantID)
	if err != nil {
		http.Error(w, "Failed to retrieve or create notification queue", http.StatusInternalServerError)
		return
	}

	// Initialize repositories
	notificationRepo := repos.NewNotificationRepository(database)
	templatesRepo := templatesrepo.NewTemplateRepository(database)
	providerRepo := providerrepos.NewProviderRepository(database)

	// Initialize services
	providerFactory := providers.NewProviderFactory(h.GenericCache, providerRepo)
	providerService := providerService.NewProviderService(providerFactory, notificationQueue, h.workerPoolManager)
	templateService := templates.NewTemplateService(templatesRepo)

	// Initialize use case
	sendNotificationUseCase := sendnotification.NewSendNotificationUseCase(providerService, templateService, providerRepo, notificationRepo, h.GenericCache)

	// Initialize controller
	sendNotificationController := sendnotification.NewSendNotificationController(sendNotificationUseCase)

	// Decode the request body
	var req sendnotification.SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute the controller method
	res, err := sendNotificationController.SendNotification(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode the response
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func NewRouter(dbManager *db.Manager, providerCache *cache.GenericCache, queueManager *queue.QueueManager, wpm *workerpool.WorkerPoolManager) *chi.Mux {
	r := chi.NewRouter()

	// Initialize handlers
	handlers := NewHandlers(dbManager, providerCache, queueManager, wpm)

	// Set up routes
	r.Post("/", handlers.SendNotification)

	// Add more routes here
	// r.Get("/another-route", handlers.AnotherHandler)

	return r
}
