package server

import (
    "getnoti.com/config"
	"github.com/go-chi/chi/v5"
    "getnoti.com/pkg/httpserver"
)

type Server struct {
    config     *config.Config
    router     *chi.Mux
    httpServer *httpserver.Server
}

func New(cfg *config.Config, r *chi.Mux) *Server {
    return &Server{
        config: cfg,
        router: r,
    }
}

func (s *Server) Start() error {
    s.httpServer = httpserver.New(s.config, s.router)
    // s.httpServer
	return nil
}

func (s *Server) Shutdown() error {
    return s.httpServer.Shutdown()
}
