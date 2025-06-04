package router

import (
	"net/http"

	"getnoti.com/internal/container"
	notificationroutes "getnoti.com/internal/notifications/infra/http"
	providerroutes "getnoti.com/internal/providers/infra/http"
	"getnoti.com/internal/server/middleware"
	"getnoti.com/internal/shared/handler"
	tenantMiddleware "getnoti.com/internal/shared/middleware"
	templateroutes "getnoti.com/internal/templates/infra/http"
	preferencesroutes "getnoti.com/internal/tenants/infra/http/preferences"
	tenantroutes "getnoti.com/internal/tenants/infra/http/tenants"
	userroutes "getnoti.com/internal/tenants/infra/http/users"
	webhookroutes "getnoti.com/internal/webhooks/infra/http"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/credentials"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/queue"
	sse "getnoti.com/pkg/sse"
	"getnoti.com/pkg/workerpool"
	"github.com/go-chi/chi/v5"
)

type Router struct {
    serviceContainer   *container.ServiceContainer
    dbManager          *db.Manager
    mainDB             db.Database
    genericCache       *cache.GenericCache
    queueManager       *queue.QueueManager
    workerPoolManager  *workerpool.WorkerPoolManager
    credentialManager  *credentials.Manager
    sseServer sse.Server
}

func New(serviceContainer *container.ServiceContainer, mainDB db.Database, dbManager *db.Manager, genericCache *cache.GenericCache, queueManager *queue.QueueManager, workerPoolManager *workerpool.WorkerPoolManager, credentialManager *credentials.Manager) *Router {
    return &Router{
        serviceContainer:  serviceContainer,
        dbManager:         dbManager,
        mainDB:            mainDB,
        genericCache:      genericCache,
        queueManager:      queueManager,
        workerPoolManager: workerPoolManager,
        credentialManager: credentialManager,
        sseServer:         sse.New(),
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
        notificationroutes.NewRouter(r.serviceContainer, r.dbManager, r.genericCache, r.queueManager, r.credentialManager, r.workerPoolManager))
    v1Router.Mount("/tenants", 
        tenantroutes.NewRouter(r.mainDB, r.dbManager, r.credentialManager))
    v1Router.With(tenantMiddleware.WithTenantID).Mount("/users", 
        userroutes.NewRouter(r.dbManager))
    v1Router.With(tenantMiddleware.WithTenantID).Mount("/templates", 
        templateroutes.NewRouter(r.dbManager))    
    v1Router.With(tenantMiddleware.WithTenantID).Mount("/providers", 
        providerroutes.NewRouter(r.dbManager))
    v1Router.With(tenantMiddleware.WithTenantID).Mount("/webhooks", 
        webhookroutes.NewRouter(r.serviceContainer, r.dbManager))    
    v1Router.With(tenantMiddleware.WithTenantID).Mount("/preferences", 
        preferencesroutes.NewRouter(handler.NewBaseHandler(r.dbManager)))

    // Add SSE endpoint for tenant (tenantMiddleware must be applied to extract tenantID)
    v1Router.With(tenantMiddleware.WithTenantID).Get("/events/stream", func(w http.ResponseWriter, req *http.Request) {
        tenantID, ok := req.Context().Value(tenantMiddleware.TenantIDKey).(string)
        if !ok || tenantID == "" {
            http.Error(w, "tenant ID required", http.StatusBadRequest)
            return
        }
        channel := "tenant_" + tenantID
        r.sseServer.ServeHTTP(w, req, channel)
    })

    router.Mount("/v1", v1Router)
}