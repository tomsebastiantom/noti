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
    "getnoti.com/pkg/credentials"
    "getnoti.com/pkg/db"
    "getnoti.com/pkg/errors"
    log "getnoti.com/pkg/logger"
    "getnoti.com/pkg/migration"
    "getnoti.com/pkg/queue"
    "getnoti.com/pkg/workerpool"
)

type App struct {
    config            *config.Config
    logger            log.Logger
    dbManager         *db.Manager
    mainDB            db.Database
    cache             *cache.GenericCache
    server            *server.Server
    queueManager      *queue.QueueManager
    workerPoolManager *workerpool.WorkerPoolManager
    credentialManager *credentials.Manager
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

 
    a.logger = log.New(a.config)
    a.logger.Info("Application initialization started",
        log.String("service", a.config.App.Name),
        log.String("version", a.config.App.Version),
        log.String("environment", a.config.Env),
    )


    a.cache = cache.NewGenericCache(1e7, 1<<30, 64)
    a.logger.Info("Cache initialized successfully")

  
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

    
    if err := migration.Migrate(a.config.Database.DSN, a.config.Database.Type, true); err != nil {
        appErr := errors.DatabaseError(ctx, "migration", err)
        a.logger.LogErrorContext(ctx, appErr)
        return appErr
    }
    a.logger.InfoContext(ctx, "Database migrations completed successfully")


    a.credentialManager, err = credentials.NewManager(a.config, a.logger, a.mainDB)
    if err != nil {
        appErr := errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("credential_manager_initialization").
            WithCause(err).
            WithMessage("Failed to initialize credential manager").
            Build()
        a.logger.LogErrorContext(ctx, appErr)
        return appErr
    }
    
    storageInfo := a.credentialManager.GetStorageInfo()
    a.logger.InfoContext(ctx, "Credential manager initialized successfully",
        log.Bool("vault_enabled", storageInfo["vault_enabled"].(bool)),
        log.String("storage_type", storageInfo["storage_type"].(string)))

   
    a.dbManager = db.NewManager(a.cache, a.config, a.logger, a.mainDB)
    a.dbManager.SetCredentialService(a.credentialManager)
    a.logger.Info("Database manager initialized successfully")

    
    a.queueManager = queue.NewQueueManager(queue.Config(a.config.Queue), a.logger)
    a.logger.Info("Queue manager initialized successfully")

  
    a.workerPoolManager = workerpool.NewWorkerPoolManager(a.logger)
    a.logger.Info("Worker pool manager initialized successfully")


    r := router.New(a.mainDB, a.dbManager, a.cache, a.queueManager, a.workerPoolManager, a.credentialManager)
    a.server = server.New(a.config, r.Handler(), a.logger)

    a.logger.Info("Application initialization completed successfully")
    return nil
}

func (a *App) Run() error {
    ctx := context.Background()
    
    storageInfo := a.credentialManager.GetStorageInfo()
    a.logger.InfoContext(ctx, "Starting notification service",
        log.String("port", a.config.HTTP.Port),
        log.String("environment", a.config.Env),
        log.String("version", a.config.App.Version),
        log.Bool("vault_enabled", storageInfo["vault_enabled"].(bool)),
    )

   
    errChan := make(chan error, 1)
    go func() {
        if err := a.server.Start(); err != nil {
            errChan <- err
        }
    }()

  
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

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
            log.String("signal", sig.String()),
        )
        return a.Cleanup()
    }
}

func (a *App) Cleanup() error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    a.logger.InfoContext(ctx, "Starting graceful shutdown")

    var cleanupErrors []error


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


    if a.queueManager != nil {
        if err := a.queueManager.Close(); err != nil {
            appErr := errors.QueueConnectionError(ctx, err)
            a.logger.LogErrorContext(ctx, appErr)
            cleanupErrors = append(cleanupErrors, appErr)
        } else {
            a.logger.InfoContext(ctx, "Queue manager closed successfully")
        }
    }


    if a.dbManager != nil {
        if err := a.dbManager.Close(); err != nil {
            appErr := errors.DatabaseError(ctx, "manager_close", err)
            a.logger.LogErrorContext(ctx, appErr)
            cleanupErrors = append(cleanupErrors, appErr)
        }
    }


    if a.mainDB != nil {
        if err := a.mainDB.Close(); err != nil {
            appErr := errors.DatabaseError(ctx, "connection_close", err)
            a.logger.LogErrorContext(ctx, appErr)
            cleanupErrors = append(cleanupErrors, appErr)
        } else {
            a.logger.InfoContext(ctx, "Database connections closed successfully")
        }
    }

 
    if len(cleanupErrors) > 0 {
        a.logger.ErrorContext(ctx, "Cleanup completed with errors",
            log.Int("error_count", len(cleanupErrors)),
        )
        return cleanupErrors[0]
    }

    a.logger.InfoContext(ctx, "Graceful shutdown completed successfully")
    return nil
}

func (a *App) HealthCheck(ctx context.Context) error {

    if err := a.mainDB.Ping(ctx); err != nil {
        return errors.DatabaseError(ctx, "health_check", err)
    }


    if !a.credentialManager.IsHealthy() {
        return errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("health_check").
            WithMessage("Credential manager is not healthy").
            Build()
    }


    if !a.queueManager.IsHealthy() {
        return errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("health_check").
            WithMessage("Queue manager is not healthy").
            Build()
    }


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

func (a *App) Shutdown(ctx context.Context) error {
    a.logger.InfoContext(ctx, "Initiating graceful shutdown")
    
    if err := a.Cleanup(); err != nil {
        a.logger.ErrorContext(ctx, "Shutdown completed with errors")
        return err
    }
    
    a.logger.InfoContext(ctx, "Shutdown completed successfully")
    return nil
}