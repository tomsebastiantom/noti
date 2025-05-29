package server

import (
    "getnoti.com/config"
    "getnoti.com/pkg/httpserver"
    "getnoti.com/pkg/logger"
    "github.com/go-chi/chi/v5"
)

type Server struct {
    config     *config.Config
    router     *chi.Mux
    httpServer *httpserver.Server
    logger     logger.Logger  // ✅ Fixed: Use interface, not pointer
}

func New(cfg *config.Config, r *chi.Mux, l logger.Logger) *Server {  // ✅ Fixed: Removed extra space
    return &Server{
        config: cfg,
        router: r,
        logger: l,  // ✅ Fixed: Removed colon syntax error
    }
}

func (s *Server) Start() error {
    s.logger.Info("Starting HTTP server",
        logger.Field{Key: "port", Value: s.config.HTTP.Port},
        logger.Field{Key: "environment", Value: s.config.Env},
    )

    var err error
    s.httpServer, err = httpserver.New(s.config, s.router, s.logger)

    if err != nil {
        // ✅ Fixed: Use structured logging
        s.logger.Error("Failed to start HTTP server",
            logger.Field{Key: "error", Value: err.Error()},
        )
        return err
    }

    s.logger.Info("HTTP server started successfully",
        logger.Field{Key: "port", Value: s.config.HTTP.Port},
    )

    return nil
}

func (s *Server) Shutdown() error {
    s.logger.Info("Shutting down HTTP server")

    err := s.httpServer.Shutdown()
    if err != nil {
        s.logger.Error("Failed to shutdown HTTP server gracefully",
            logger.Field{Key: "error", Value: err.Error()},
        )
        return err
    }

    s.logger.Info("HTTP server shut down successfully")
    return nil
}