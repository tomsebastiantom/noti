
package notificationroutes

import (
	"context"
	"encoding/json"
	"net/http"

	
	"getnoti.com/internal/notifications/repos/implementations"
	"getnoti.com/internal/notifications/usecases/send_notification"
	"github.com/go-chi/chi/v5"
	"getnoti.com/pkg/db"

	"getnoti.com/internal/providers/infra/providers"
	providerService"getnoti.com/internal/providers/services"

    "getnoti.com/internal/tenants/services"
 

	
	pg "getnoti.com/internal/tenants/repos/implementations"

	
)


    
    
   

func NewRouter(database db.Database) *chi.Mux {
	r := chi.NewRouter()

    // Initialize repositories
    notificationRepo := postgres.NewPostgresNotificationRepository(database)
    tenantRepo := pg.NewPostgresTenantRepository(database)

    // Initialize services
    tenantService := tenants.NewTenantService(tenantRepo)
    providerFactory := providers.NewProviderFactory() // Assuming you have a factory initialization
    providerService := providerService.NewProviderService(providerFactory)

    // Initialize use case
    sendNotificationUseCase := sendnotification.NewSendNotificationUseCase(tenantService, providerService, notificationRepo)

    // Initialize controller
    sendNotificationController := sendnotification.NewSendNotificationController(sendNotificationUseCase)

    // Set up routes
    r.Post("/", commonHandler(sendNotificationController.SendNotification))

    return r

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
