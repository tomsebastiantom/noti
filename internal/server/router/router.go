package router

import (
	"getnoti.com/internal/notifications/infra/http"
	providerroutes "getnoti.com/internal/providers/infra/http"
	"getnoti.com/internal/server/middleware"

	"getnoti.com/internal/templates/infra/http"
	"getnoti.com/internal/tenants/infra/http/tenants"
	"getnoti.com/internal/tenants/infra/http/users"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/vault"
	"github.com/go-chi/chi/v5"
	"getnoti.com/pkg/workerpool"
)

type Router struct {
	dbManager    *db.Manager
	mainDB       db.Database
	vaultCfg     *vault.VaultConfig
	genericCache *cache.GenericCache
	queueManager *queue.QueueManager
	workerPoolManager *workerpool.WorkerPoolManager
}

func New(mainDB db.Database, dbManager *db.Manager, vaultCfg *vault.VaultConfig, genericCache *cache.GenericCache, queueManager *queue.QueueManager,workerPoolManager *workerpool.WorkerPoolManager) *Router {
	return &Router{dbManager: dbManager,
		mainDB:       mainDB,
		vaultCfg:     vaultCfg,
		genericCache: genericCache,
		queueManager:        queueManager,
		workerPoolManager :workerPoolManager,
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

	v1Router.Mount("/notifications", notificationroutes.NewRouter(r.dbManager, r.genericCache, r.queueManager,r.workerPoolManager))
	v1Router.Mount("/tenants", tenantroutes.NewRouter(r.mainDB, r.dbManager, r.vaultCfg))
	v1Router.Mount("/users", userroutes.NewRouter(r.dbManager))
	v1Router.Mount("/templates", templateroutes.NewRouter(r.dbManager))
	v1Router.Mount("/providers",providerroutes.NewRouter(r.dbManager))

	router.Mount("/v1", v1Router)
}
