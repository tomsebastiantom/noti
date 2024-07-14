package server

import (
    "getnoti.com/config"
    "getnoti.com/internal/server/router"
    "getnoti.com/pkg/httpserver"
)

type Server struct {
    config     *config.Config
    router     *router.Router
    httpServer *httpserver.Server
}

func New(cfg *config.Config, r *router.Router) *Server {
    return &Server{
        config: cfg,
        router: r,
    }
}

func (s *Server) Start() error {
    s.httpServer = httpserver.New(s.config, s.router.Handler())
    // s.httpServer
	return nil
}

func (s *Server) Shutdown() error {
    return s.httpServer.Shutdown()
}
