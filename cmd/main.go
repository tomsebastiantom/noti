package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"getnoti.com/config"
	"getnoti.com/internal/notifications/infra/http"
	custom "getnoti.com/internal/shared/middleware"
	"getnoti.com/internal/templates/infra/http"
	"getnoti.com/internal/tenants/infra/http/tenants"
	"getnoti.com/internal/tenants/infra/http/users"
	"getnoti.com/pkg/cache"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/httpserver"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/vault"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()

	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// fmt.Printf("App Name: %s\n", cfg.App.Name)
	// fmt.Printf("App Version: %s\n", cfg.App.Version)
	// fmt.Printf("HTTP Port: %s\n", cfg.HTTP.Port)
	// fmt.Printf("Log Level: %s\n", cfg.Log.Level)
	// fmt.Printf("Postgres PoolMax: %d\n", cfg.Database.Postgres.PoolMax)
	// fmt.Printf("Postgres URL: %s\n", cfg.Database.Postgres.DSN)
	// fmt.Printf("RabbitMQ Server Exchange: %s\n", cfg.RMQ.ServerExchange)
	// fmt.Printf("RabbitMQ Client Exchange: %s\n", cfg.RMQ.ClientExchange)
	// fmt.Printf("RabbitMQ URL: %s\n", cfg.RMQ.URL)

	// Initialize the main database
	mainDB, err := db.NewDatabaseFactory((*db.DatabaseConfig)(&cfg.Database))
	if err != nil {
		fmt.Printf("Failed to initialize main database: %v\n", err)
		os.Exit(1)
	}
	defer mainDB.(*db.SQLDatabase).Close()

	// Initialize cache
	genericCache := cache.NewGenericCache(1e7, 1<<30, 64)

	// migrate.Migrate(cfg.Database.DSN)

	// Initialize logger

	log := logger.New(cfg)

	// Initialize the database manager
	dbManager := db.NewManager(genericCache, (*vault.VaultConfig)(&cfg.Vault))

	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	// defer database.Close()
	// Create main router
	router := chi.NewRouter()

	// Use common middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(custom.TenantID)

	// Create version-specific routers
	v1Router := chi.NewRouter()

	// Mount routes
	notificationRouter := notificationroutes.NewRouter(dbManager)
	tenantRouter := tenantroutes.NewRouter(mainDB,dbManager,(*vault.VaultConfig)(&cfg.Vault))
	userRouter := userroutes.NewRouter(dbManager)
	templateRouter := templateroutes.NewRouter(dbManager)

	v1Router.Mount("/notifications", notificationRouter)
	v1Router.Mount("/users", userRouter)
	v1Router.Mount("/tenants", tenantRouter)
	v1Router.Mount("/templates", templateRouter)

	router.Mount("/v1", v1Router)

	// Create HTTP server
	httpServer := httpserver.New(cfg, router)

	// Start the server
	log.Info("Server started on port " + cfg.HTTP.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Server is shutting down...")

	if err := httpServer.Shutdown(); err != nil {
		log.Error(fmt.Errorf("server shutdown: %w", err))
	}

	log.Info("Server exited properly")
}
