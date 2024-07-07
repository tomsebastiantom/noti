package providerroutes

import (
    "context"
    "encoding/json"
    "net/http"

    "getnoti.com/internal/providers/infra/providers"
    "getnoti.com/internal/providers/usecases/send_notification"
    "github.com/go-chi/chi/v5"
    "getnoti.com/pkg/db"
)

func NewRouter(database db.Database) *chi.Mux {
    r := chi.NewRouter()

    // Initialize provider
    twilioProvider := providers.NewTwilioProvider()

    // Initialize use case
    sendNotificationUseCase := sendnotification.NewSendNotificationUseCase(twilioProvider)

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
        case func(context.Context, providerdto.SendNotificationRequest) providerdto.SendNotificationResponse:
            res = h(ctx, req.(providerdto.SendNotificationRequest))

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
