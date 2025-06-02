package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"getnoti.com/config"
	"getnoti.com/internal/container"
	"getnoti.com/internal/server"
	"getnoti.com/internal/server/router"
	"getnoti.com/pkg/errors"
	log "getnoti.com/pkg/logger"
	"getnoti.com/pkg/migration"
)

type App struct {
    config           *config.Config
    logger           log.Logger
    serviceContainer *container.ServiceContainer
    server           *server.Server
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

    // Initialize logger first
    a.logger = log.New(a.config)
    a.logger.Info("Application initialization started",
        log.String("service", a.config.App.Name),
        log.String("version", a.config.App.Version),
        log.String("environment", a.config.Env),
    )

    // Run migrations before initializing service container
    if err := migration.Migrate(a.config.Database.DSN, a.config.Database.Type, true); err != nil {
        appErr := errors.DatabaseError(ctx, "migration", err)
        a.logger.LogErrorContext(ctx, appErr)
        return appErr
    }
    a.logger.InfoContext(ctx, "Database migrations completed successfully")

    // Initialize service container (replaces all manual dependency creation)
    a.serviceContainer, err = container.NewServiceContainer(a.config, a.logger)
    if err != nil {
        appErr := errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("container_initialization").
            WithCause(err).
            WithMessage("Failed to initialize service container").
            Build()
        a.logger.LogErrorContext(ctx, appErr)
        return appErr
    }
    a.logger.Info("Service container initialized successfully")

    // Get infrastructure for router
    infrastructure, err := a.serviceContainer.GetInfrastructure()
    if err != nil {
        appErr := errors.New(errors.ErrCodeInternal).
            WithContext(ctx).
            WithOperation("infrastructure_get").
            WithCause(err).
            WithMessage("Failed to get infrastructure components").
            Build()
        a.logger.LogErrorContext(ctx, appErr)
        return appErr
    }    // Initialize router with infrastructure and service container
    r := router.New(
        a.serviceContainer,
        infrastructure.MainDB,
        infrastructure.DBManager,
        infrastructure.Cache,
        infrastructure.QueueManager,
        infrastructure.WorkerPoolManager,
        infrastructure.CredentialManager,
    )
    a.server = server.New(a.config, r.Handler(), a.logger)

    a.logger.Info("Application initialization completed successfully")
    return nil
}

func (a *App) Run() error {
    ctx := context.Background()
    
    a.logger.InfoContext(ctx, "Starting notification service",
        log.String("port", a.config.HTTP.Port),
        log.String("environment", a.config.Env),
        log.String("version", a.config.App.Version),
    )

    // Start server in goroutine
    errChan := make(chan error, 1)
    go func() {
        if err := a.server.Start(); err != nil {
            errChan <- err
        }
    }()

    // Setup signal handling
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

    // Shutdown server first
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

    // Cleanup service container (handles all infrastructure cleanup)
    if a.serviceContainer != nil {
        if err := a.serviceContainer.Cleanup(ctx); err != nil {
            appErr := errors.New(errors.ErrCodeInternal).
                WithContext(ctx).
                WithOperation("container_cleanup").
                WithCause(err).
                WithMessage("Failed to cleanup service container").
                Build()
            a.logger.LogErrorContext(ctx, appErr)
            cleanupErrors = append(cleanupErrors, appErr)
        } else {
            a.logger.InfoContext(ctx, "Service container cleaned up successfully")
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
    // Delegate health check to service container
    if err := a.serviceContainer.HealthCheck(ctx); err != nil {
        return err
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