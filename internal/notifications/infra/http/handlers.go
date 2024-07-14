package notificationroutes

import (
    "context"
    "encoding/json"
    "net/http"
   	"getnoti.com/internal/notifications/repos/implementations"
    "getnoti.com/internal/notifications/usecases/send_notification"
    "getnoti.com/pkg/db"
    "getnoti.com/internal/providers/infra/providers"
    providerService "getnoti.com/internal/providers/services"
    tenants "getnoti.com/internal/tenants/services"
    repos "getnoti.com/internal/tenants/repos/implementations"
    custom "getnoti.com/internal/shared/middleware"
)

type Handlers struct {
    DBManager *db.Manager
}

func NewHandlers(dbManager *db.Manager) *Handlers {
    return &Handlers{
        DBManager: dbManager,
    }
}

func (h *Handlers) SendNotification(w http.ResponseWriter, r *http.Request) {
    tenantID := r.Context().Value(custom.TenantIDKey).(string)

    // Retrieve the database connection
    database, err := h.DBManager.GetDatabaseConnection(tenantID)
    if err != nil {
        http.Error(w, "Failed to retrieve database connection", http.StatusInternalServerError)
        return
    }

    // Initialize repositories
    notificationRepo := postgres.NewPostgresNotificationRepository(database)
    // tenantRepo := repos.NewTenantRepository(database)
	tenantRepo := repos.NewTenantPreferenceRepository(database)
    // Initialize services
    tenantService := tenants.NewTenantService(tenantRepo)
    providerFactory := providers.NewProviderFactory()
    providerService := providerService.NewProviderService(providerFactory)

    // Initialize use case
    sendNotificationUseCase := sendnotification.NewSendNotificationUseCase(tenantService, providerService, notificationRepo)

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
