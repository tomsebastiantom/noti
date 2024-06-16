//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.
package internal

import (
    "github.com/gin-gonic/gin"
    "github.com/google/wire"
    "workerhive.com/api/config"
    "workerhive.com/api/pkg/httpserver"
    "workerhive.com/api/pkg/logger"
    "workerhive.com/api/pkg/postgres"
)

// Define a provider set with the necessary providers.
var providerSet = wire.NewSet(
    postgres.NewOrGetSingleton,
    logger.New,
    httpserver.New,
    NewRouter, // Add this line
    LoadConfig, // Add this line
)

// Define a basic provider for the Gin router.
func NewRouter() *gin.Engine {
    router := gin.Default()
    return router
}

// LoadConfig loads the configuration using koanf.
func LoadConfig() (*config.Config, error) {
    return config.LoadConfig("config.yaml")
}

// Define the initialization functions.
func InitializeConfig() (*config.Config, error) {
    wire.Build(LoadConfig)
    return &config.Config{}, nil
}

func InitializePostgresConnection() (*postgres.Postgres, error) {
    wire.Build(providerSet)
    return &postgres.Postgres{}, nil
}

func InitializeLogger() (*logger.Logger, error) {
    wire.Build(providerSet)
    return &logger.Logger{}, nil
}

func InitializeNewRouter() (*gin.Engine, error) {
    wire.Build(providerSet)
    return &gin.Engine{}, nil
}

func InitializeNewHttpServer() (*httpserver.Server, error) {
    wire.Build(providerSet)
    return &httpserver.Server{}, nil
}
