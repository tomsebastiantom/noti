package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"getnoti.com/config"
	"getnoti.com/internal/notifications/infra/http"
	"getnoti.com/pkg/httpserver"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/postgres"
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
	// fmt.Printf("Postgres PoolMax: %d\n", cfg.PG.PoolMax)
	// fmt.Printf("Postgres URL: %s\n", cfg.PG.URL)
	// fmt.Printf("RabbitMQ Server Exchange: %s\n", cfg.RMQ.ServerExchange)
	// fmt.Printf("RabbitMQ Client Exchange: %s\n", cfg.RMQ.ClientExchange)
	// fmt.Printf("RabbitMQ URL: %s\n", cfg.RMQ.URL)

	// Initialize logger
	log := logger.New(cfg)

	// Initialize database connection pool
	db := postgres.NewOrGetSingleton(cfg)
	defer db.Close()

	// Create main router
	router := chi.NewRouter()

	// Use common middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Mount notification routes
	notificationRouter := notificationroutes.NewRouter(db.Pool)
	router.Mount("/notifications", notificationRouter)

	// Mount other domain routers here as needed
	// userRouter := users.NewRouter(db.Pool)
	// router.Mount("/users", userRouter)


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
