package userroutes

import (
	"context"
	"encoding/json"
	"net/http"

	"getnoti.com/internal/notifications/usecases/sendnotification"
	"github.com/go-chi/chi/v5"
)

func NewRouter(
	sendNotificationController *sendnotification.SendNotificationController,

) *chi.Mux {
	r := chi.NewRouter()

	// Middleware to extract IDs from URL parameters and add them to the request context
	// r.Use(middleware.ExtractIDMiddleware)

	// User routes
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
