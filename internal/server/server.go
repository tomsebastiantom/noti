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
    logger          *logger.Logger
}

func New(cfg *config.Config, r *chi.Mux,l  *logger.Logger) *Server {
    return &Server{
        config: cfg,
        router: r,
        logger  :l,
    }
}

func (s *Server) Start() error {
    var err error
    s.httpServer,err = httpserver.New(s.config, s.router,s.logger)
 
    if err!=nil {
        s.logger.Error("Failed to Start HTTP server "  + err.Error())
        return err
    }
	return nil
}

func (s *Server) Shutdown() error {
    return s.httpServer.Shutdown()
}
