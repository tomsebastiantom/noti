package router

import (
    "getnoti.com/internal/server/middleware"
 "getnoti.com/internal/notifications/infra/http"
	// customMiddleware"getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/templates/infra/http"
	// "getnoti.com/internal/tenants/infra/http/tenants"
	"getnoti.com/internal/tenants/infra/http/users"
    "getnoti.com/pkg/db"
    "github.com/go-chi/chi/v5"
)

type Router struct {
    dbManager *db.Manager
}

func New(dbManager *db.Manager) *Router {
    return &Router{dbManager: dbManager}
}

func (r *Router) Handler() *chi.Mux {
    router := chi.NewRouter()

    // Apply middleware
    middleware.Apply(router)

    // Mount routes
    r.mountV1Routes(router)

    return router
}

func (r *Router) mountV1Routes(router chi.Router) {
    v1Router := chi.NewRouter()

    v1Router.Mount("/notifications", notificationroutes.NewRouter(r.dbManager))
    // v1Router.Mount("/users", tenantroutes.NewRouter(r.dbManager))
    v1Router.Mount("/tenants", userroutes.NewRouter(r.dbManager))
    v1Router.Mount("/templates", templateroutes.NewRouter(r.dbManager))

    router.Mount("/v1", v1Router)
}
