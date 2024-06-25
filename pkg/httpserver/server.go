package httpserver

import (
	"context"
	"net"
	"net/http"
	"time"

	"getnoti.com/config"
	"github.com/go-chi/chi/v5"
)

const (
	_defaultReadTimeout     = 5 * time.Second
	_defaultWriteTimeout    = 5 * time.Second
	_defaultAddr            = ":80"
	_defaultShutdownTimeout = 3 * time.Second
)

// Server -.
type Server struct {
	server          *http.Server
	notify          chan error
	shutdownTimeout time.Duration
	Router          *chi.Mux
}

// New creates a new HTTP server.
func New(cfg *config.Config, router *chi.Mux) *Server {
	server := prepareHttpServer(cfg, router)
	server.start()

	return server
}

func prepareHttpServer(cfg *config.Config, router *chi.Mux) *Server {
	httpServer := &http.Server{
		Handler:      router,
		ReadTimeout:  _defaultReadTimeout,
		WriteTimeout: _defaultWriteTimeout,
		Addr:         _defaultAddr,
	}
	httpServer.Addr = net.JoinHostPort("", cfg.HTTP.Port)

	s := &Server{
		server:          httpServer,
		notify:          make(chan error, 1),
		shutdownTimeout: _defaultShutdownTimeout,
		Router:          router,
	}
	return s
}

func (s *Server) start() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

// Notify returns a channel to notify when the server is closed.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
