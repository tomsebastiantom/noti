package notificationroutes

import (
	"context"
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
	// Initialize repositories
	notificationRepo := repos.NewNotificationRepository(database)

	templatesRepo := templatesrepo.NewTemplateRepository(database)

	providerRepo := providerrepos.NewProviderRepository(database)
	// Initialize services

	providerFactory := providers.NewProviderFactory(h.GenericCache, providerRepo)
	providerService := providerService.NewProviderService(providerFactory, notificationQueue, h.workerPoolManager)
	templateService := templates.NewTemplateService(templatesRepo)
	// Initialize use case
	sendNotificationUseCase := sendnotification.NewSendNotificationUseCase(providerService, templateService,providerRepo, notificationRepo, h.GenericCache)

	// Initialize controller
	sendNotificationController := sendnotification.NewSendNotificationController(sendNotificationUseCase)

	// Handle the request
	commonHandler(sendNotificationController.SendNotification)(w, r)
}

// commonHandler is a generic HTTP handler function that handles requests and responses for different controllers.
func commonHandler(handlerFunc interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Decode the request body into the appropriate request type
		var req interface{}
		if r.Method != http.MethodGet && r.Method != http.MethodDelete {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
		}

		// Call the handler function with the context and request
		var res interface{}
		switch h := handlerFunc.(type) {
		case func(context.Context, sendnotification.SendNotificationRequest) sendnotification.SendNotificationResponse:
			res = h(ctx, req.(sendnotification.SendNotificationRequest))

		default:
			http.Error(w, "Unsupported handler function", http.StatusInternalServerError)
			return
		}

		// Encode the response and write it to the response writer
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
