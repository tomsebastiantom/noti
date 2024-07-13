package notificationroutes

import (
	"getnoti.com/pkg/db"
	"github.com/go-chi/chi/v5"
)

func NewRouter(dbManager *db.Manager) *chi.Mux {
	r := chi.NewRouter()

	// Initialize handlers
	handlers := NewHandlers(dbManager)

	// Set up routes
	r.Post("/", handlers.SendNotification)

	// Add more routes here
	// r.Get("/another-route", handlers.AnotherHandler)

	return r
}
