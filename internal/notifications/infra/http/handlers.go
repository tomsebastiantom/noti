package notificationroutes

import (
	"context"
	"encoding/json"
	"net/http"

	"getnoti.com/internal/notifications/repos/implementations"
	"getnoti.com/internal/notifications/usecases/send_notification"
	"getnoti.com/internal/providers/infra/providers"
	providerService "getnoti.com/internal/providers/services"
	"getnoti.com/internal/shared/middleware"
	templates "getnoti.com/internal/templates/services"
	tenantrepos "getnoti.com/internal/tenants/repos/implementations"
	templatesrepo "getnoti.com/internal/templates/repos/implementations"
	tenants "getnoti.com/internal/tenants/services"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/db"
)

type Handlers struct {
	DBManager *db.Manager
	GenericCache  *cache.GenericCache
}

func NewHandlers(dbManager *db.Manager,genericCache *cache.GenericCache) *Handlers {
	return &Handlers{
		DBManager: dbManager,
		GenericCache : genericCache ,
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

	// Initialize repositories
	notificationRepo := repos.NewNotificationRepository(database)
	// tenantRepo := repos.NewTenantRepository(database)
	tenantRepo := tenantrepos.NewTenantPreferenceRepository(database)

	templatesRepo := templatesrepo.NewTemplateRepository(database)
	// Initialize services
	tenantService := tenants.NewTenantService(tenantRepo)
	providerFactory := providers.NewProviderFactory(h.GenericCache)
	providerService := providerService.NewProviderService(providerFactory)
    templateService := templates.NewTemplateService(templatesRepo)
	// Initialize use case
	sendNotificationUseCase := sendnotification.NewSendNotificationUseCase(tenantService, providerService,templateService, notificationRepo,h.GenericCache)

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
