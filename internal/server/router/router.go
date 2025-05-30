// package router

// import (
// 	"getnoti.com/internal/notifications/infra/http"
// 	providerroutes "getnoti.com/internal/providers/infra/http"
// 	"getnoti.com/internal/server/middleware"
// 	tenantMiddleware "getnoti.com/internal/shared/middleware"
// 	"getnoti.com/internal/templates/infra/http"
// 	"getnoti.com/internal/tenants/infra/http/tenants"
// 	"getnoti.com/internal/tenants/infra/http/users"
// 	"getnoti.com/pkg/cache"
// 	"getnoti.com/pkg/db"
// 	"getnoti.com/pkg/queue"
// 	"getnoti.com/pkg/workerpool"
// 	"github.com/go-chi/chi/v5"
// )

// type Router struct {
// 	dbManager         *db.Manager
// 	mainDB            db.Database
// 	genericCache      *cache.GenericCache
// 	queueManager      *queue.QueueManager
// 	workerPoolManager *workerpool.WorkerPoolManager
// }

// func New(mainDB db.Database, dbManager *db.Manager, genericCache *cache.GenericCache, queueManager *queue.QueueManager, workerPoolManager *workerpool.WorkerPoolManager) *Router {
// 	return &Router{dbManager: dbManager,
// 		mainDB:            mainDB,
// 		genericCache:      genericCache,
// 		queueManager:      queueManager,
// 		workerPoolManager: workerPoolManager,
// 	}
// }

// func (r *Router) Handler() *chi.Mux {

// 	router := chi.NewRouter()

// 	// Apply middleware
// 	middleware.Apply(router)

// 	// Mount routes
// 	r.mountV1Routes(router)

// 	return router
// }

// func (r *Router) mountV1Routes(router chi.Router) {
// 	v1Router := chi.NewRouter()

// 	v1Router.With(tenantMiddleware.WithTenantID).Mount("/notifications", notificationroutes.NewRouter(r.dbManager, r.genericCache, r.queueManager, r.workerPoolManager))
// 	v1Router.Mount("/tenants", tenantroutes.NewRouter(r.mainDB, r.dbManager))
// 	v1Router.With(tenantMiddleware.WithTenantID).Mount("/users", userroutes.NewRouter(r.dbManager))
// 	v1Router.With(tenantMiddleware.WithTenantID).Mount("/templates", templateroutes.NewRouter(r.dbManager))
// 	v1Router.With(tenantMiddleware.WithTenantID).Mount("/providers", providerroutes.NewRouter(r.dbManager))

// 	router.Mount("/v1", v1Router)
// }

package router

import (
    "getnoti.com/internal/notifications/infra/http"
    providerroutes "getnoti.com/internal/providers/infra/http"
    "getnoti.com/internal/server/middleware"
    tenantMiddleware "getnoti.com/internal/shared/middleware"
    templateroutes "getnoti.com/internal/templates/infra/http"
    "getnoti.com/internal/tenants/infra/http/tenants"
    "getnoti.com/internal/tenants/infra/http/users"
    "getnoti.com/pkg/cache"
    "getnoti.com/pkg/credentials"
    "getnoti.com/pkg/db"
    "getnoti.com/pkg/queue"
    "getnoti.com/pkg/workerpool"
    "github.com/go-chi/chi/v5"
)

type Router struct {
    dbManager          *db.Manager
    mainDB             db.Database
    genericCache       *cache.GenericCache
    queueManager       *queue.QueueManager
    workerPoolManager  *workerpool.WorkerPoolManager
    credentialManager  *credentials.Manager
}

func New(mainDB db.Database, dbManager *db.Manager, genericCache *cache.GenericCache, queueManager *queue.QueueManager, workerPoolManager *workerpool.WorkerPoolManager, credentialManager *credentials.Manager) *Router {
    return &Router{
        dbManager:         dbManager,
        mainDB:            mainDB,
        genericCache:      genericCache,
        queueManager:      queueManager,
        workerPoolManager: workerPoolManager,
        credentialManager: credentialManager,
    }
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

    v1Router.With(tenantMiddleware.WithTenantID).Mount("/notifications", 
        notificationroutes.NewRouter(r.dbManager, r.genericCache, r.queueManager,r.credentialManager, r.workerPoolManager))
    v1Router.Mount("/tenants", 
        tenantroutes.NewRouter(r.mainDB, r.dbManager)) // Keep existing signature
    v1Router.With(tenantMiddleware.WithTenantID).Mount("/users", 
        userroutes.NewRouter(r.dbManager))
    v1Router.With(tenantMiddleware.WithTenantID).Mount("/templates", 
        templateroutes.NewRouter(r.dbManager))
    v1Router.With(tenantMiddleware.WithTenantID).Mount("/providers", 
        providerroutes.NewRouter(r.dbManager)) // Keep existing signature

    router.Mount("/v1", v1Router)
}