package app

import (
	"fmt"

	"getnoti.com/config"
	"getnoti.com/internal/server"
	"getnoti.com/internal/server/router"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/vault"
	"getnoti.com/pkg/workerpool"
)

type App struct {
	config            *config.Config
	logger            *logger.Logger
	db                *db.Manager
	mainDB            db.Database
	cache             *cache.GenericCache
	server            *server.Server
	queueManager      *queue.QueueManager
	workerPoolManager *workerpool.WorkerPoolManager
}

func New(cfg *config.Config) (*App, error) {
	app := &App{
		config: cfg,
	}

	if err := app.initialize(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) initialize() error {
	var err error

	// Initialize logger
	a.logger = logger.New(a.config)

	// Initialize cache
	a.cache = cache.NewGenericCache(1e7, 1<<30, 64)

	// Initialize database manager
	a.db = db.NewManager(a.cache, (*vault.VaultConfig)(&a.config.Vault))
	if err != nil {
		return fmt.Errorf("failed to initialize database manager: %w", err)
	}

	// Initialize main database
	a.mainDB, err = db.NewDatabaseFactory((*db.DatabaseConfig)(&a.config.Database))

	if err != nil {
		return fmt.Errorf("failed to initialize main database: %w", err)
	}

	// Initialize main queueManager
	a.queueManager = queue.NewQueueManager(queue.Config(a.config.Queue), a.logger)
    //Initialize WorkerPoolManager
	a.workerPoolManager = workerpool.NewWorkerPoolManager(*a.logger)

	//Initialize router
	r := router.New(a.mainDB, a.db, (*vault.VaultConfig)(&a.config.Vault), a.cache, a.queueManager, a.workerPoolManager)

	//Initialize server
	a.server = server.New(a.config, r.Handler())

	return nil
}

func (a *App) Run() error {
	a.logger.Info("Server starting on port " + a.config.HTTP.Port)
	return a.server.Start()
}

func (a *App) Cleanup() {
	a.logger.Info("Cleaning up resources...")
	// if a.db != nil {
	// 	Close database connections
	// 	Note: You might need to implement a Close method in your db.Manager
	// 	a.db.Close()
	// }
	if a.server != nil {
		if err := a.server.Shutdown(); err != nil {
			a.logger.Error(fmt.Errorf("server shutdown error: %w", err))
		}
	}
	a.logger.Info("Cleanup completed")
}
