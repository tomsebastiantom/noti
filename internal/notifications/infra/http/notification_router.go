package notificationroutes

import (
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/cache"
	"github.com/go-chi/chi/v5"
)

func NewRouter(dbManager *db.Manager,providerCache *cache.GenericCache) *chi.Mux {
	r := chi.NewRouter()

	// Initialize handlers
	handlers := NewHandlers(dbManager,providerCache)

	// Set up routes
	r.Post("/", handlers.SendNotification)

	// Add more routes here
	// r.Get("/another-route", handlers.AnotherHandler)

	return r
}
