package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"getnoti.com/config"
	"getnoti.com/internal/server"
	"getnoti.com/internal/server/router"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/errors"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/migration"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/vault"
	"getnoti.com/pkg/workerpool"
)

type App struct {
    config            *config.Config
    logger            logger.Logger
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
    ctx := context.Background()

    // ✅ Initialize logger first
    a.logger = logger.New(a.config)
    a.logger.Info("Application initialization started",
        a.logger.String("service", a.config.App.Name),
        a.logger.String("version", a.config.App.Version),
        a.logger.String("environment", a.config.Env),
    )

    // ✅ Initialize cache
    a.cache = cache.NewGenericCache(1e7, 1<<30, 64)
    a.logger.Info("Cache initialized successfully")

    // ✅ Initialize database manager
    a.db = db.NewManager(a.cache, (*vault.VaultConfig)(&a.config.Vault), a.config, a.logger)

    // ✅ Initialize main database with proper error handling
    a.mainDB, err = db.NewDatabaseFactory(map[string]interface{}{
        "type": a.config.Database.Type,
        "dsn":  a.config.Database.DSN,
    }, a.logger)

    if err != nil {
        appErr := errors.DatabaseConnectionError(ctx, err)
        a.logger.LogErrorContext(ctx, appErr)
        return appErr
    }
    a.logger.InfoContext(ctx, "Main database initialized successfully")

    // ✅ Run migrations with proper error handling
    if err := migration.Migrate(a.config.Database.DSN, a.config.Database.Type, true); err != nil {
        appErr := errors.DatabaseError(ctx, "migration", err)
        a.logger.LogErrorContext(ctx, appErr)
        return appErr
    }
    a.logger.InfoContext(ctx, "Database migrations completed successfully")

    // ✅ Initialize queue manager
    a.queueManager = queue.NewQueueManager(queue.Config(a.config.Queue), a.logger)
    a.logger.Info("Queue manager initialized successfully")

    // ✅ Initialize worker pool manager
    a.workerPoolManager = workerpool.NewWorkerPoolManager(a.logger)
    a.logger.Info("Worker pool manager initialized successfully")

    // ✅ Initialize vault with proper error handling
    if err := vault.Initialize((*vault.VaultConfig)(&a.config.Vault)); err != nil {
        appErr := errors.VaultError(ctx, "initialization", err)
        a.logger.LogErrorContext(ctx, appErr)
        return appErr
    }
    a.logger.InfoContext(ctx, "Vault initialized successfully")

    // ✅ Initialize server
    r := router.New(a.mainDB, a.db, a.cache, a.queueManager, a.workerPoolManager)
    a.server = server.New(a.config, r.Handler(), a.logger)

    a.logger.Info("Application initialization completed successfully")
    return nil
}

func (a *App) Run() error {
    ctx := context.Background()
    
    a.logger.InfoContext(ctx, "Starting notification service",
        a.logger.String("port", a.config.HTTP.Port),
        a.logger.String("environment", a.config.Env),
        a.logger.String("version", a.config.App.Version),
    )

    // ✅ Start server in goroutine
    errChan := make(chan error, 1)
    go func() {
        if err := a.server.Start(); err != nil {
            errChan <- err
        }
    }()

    // ✅ Listen for shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // ✅ Wait for either error or signal
    select {
    case err := <-errChan:
        appErr := errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("server_start").
            WithCause(err).
            WithMessage("Failed to start server").
            Build()
        a.logger.LogErrorContext(ctx, appErr)
        return appErr

    case sig := <-sigChan:
        a.logger.InfoContext(ctx, "Received shutdown signal",
            a.logger.String("signal", sig.String()),
        )
        return a.Cleanup()
    }
}

func (a *App) Cleanup() error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    a.logger.InfoContext(ctx, "Starting graceful shutdown")

    var cleanupErrors []error

    // ✅ Shutdown server
    if a.server != nil {
        if err := a.server.Shutdown(); err != nil {
            appErr := errors.New(errors.ErrCodeInternal).
                WithContext(ctx).
                WithOperation("server_shutdown").
                WithCause(err).
                WithMessage("Failed to shutdown server").
                Build()
            a.logger.LogErrorContext(ctx, appErr)
            cleanupErrors = append(cleanupErrors, appErr)
        } else {
            a.logger.InfoContext(ctx, "Server shut down successfully")
        }
    }

    // ✅ Shutdown worker pools
    if a.workerPoolManager != nil {
        if err := a.workerPoolManager.Shutdown(); err != nil {
            appErr := errors.New(errors.ErrCodeInternal).
                WithContext(ctx).
                WithOperation("workerpool_shutdown").
                WithCause(err).
                WithMessage("Failed to shutdown worker pools").
                Build()
            a.logger.LogErrorContext(ctx, appErr)
            cleanupErrors = append(cleanupErrors, appErr)
        } else {
            a.logger.InfoContext(ctx, "Worker pools shut down successfully")
        }
    }

    // ✅ Close queue manager
    if a.queueManager != nil {
        if err := a.queueManager.Close(); err != nil {
            appErr := errors.QueueConnectionError(ctx, err)
            a.logger.LogErrorContext(ctx, appErr)
            cleanupErrors = append(cleanupErrors, appErr)
        } else {
            a.logger.InfoContext(ctx, "Queue manager closed successfully")
        }
    }

    // ✅ Close database connections
    if a.mainDB != nil {
        if err := a.mainDB.Close(); err != nil {
            appErr := errors.DatabaseError(ctx, "connection_close", err)
            a.logger.LogErrorContext(ctx, appErr)
            cleanupErrors = append(cleanupErrors, appErr)
        } else {
            a.logger.InfoContext(ctx, "Database connections closed successfully")
        }
    }

    // ✅ Report cleanup results
    if len(cleanupErrors) > 0 {
        a.logger.ErrorContext(ctx, "Cleanup completed with errors",
            a.logger.Int("error_count", len(cleanupErrors)),
        )
        return cleanupErrors[0]
    }

    a.logger.InfoContext(ctx, "Graceful shutdown completed successfully")
    return nil
}

func (a *App) HealthCheck(ctx context.Context) error {
    // ✅ Check database health
    if err := a.mainDB.Ping(ctx); err != nil {
        return errors.DatabaseError(ctx, "health_check", err)
    }

    // ✅ Check queue manager health
    if !a.queueManager.IsHealthy() {
        return errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("health_check").
            WithMessage("Queue manager is not healthy").
            Build()
    }

    // ✅ Check worker pool manager health
    if !a.workerPoolManager.IsHealthy() {
        return errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("health_check").
            WithMessage("Worker pool manager is not healthy").
            Build()
    }

    a.logger.DebugContext(ctx, "Health check passed")
    return nil
}

// ✅ ADD GRACEFUL SHUTDOWN METHOD
func (a *App) Shutdown(ctx context.Context) error {
    a.logger.InfoContext(ctx, "Initiating graceful shutdown")
    
    // Close all components in reverse order of initialization
    if err := a.Cleanup(); err != nil {
        a.logger.ErrorContext(ctx, "Shutdown completed with errors")
        return err
    }
    
    a.logger.InfoContext(ctx, "Shutdown completed successfully")
    return nil
}