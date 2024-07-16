package notificationroutes

import (
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/queue"
	"github.com/go-chi/chi/v5"
)

func NewRouter(dbManager *db.Manager, providerCache *cache.GenericCache, queueManager *queue.QueueManager) *chi.Mux {
	r := chi.NewRouter()

	// Initialize handlers
	handlers := NewHandlers(dbManager, providerCache, queueManager)

	// Set up routes
	r.Post("/", handlers.SendNotification)

	// Add more routes here
	// r.Get("/another-route", handlers.AnotherHandler)

	return r
}
