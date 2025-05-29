package httpserver

import (
	"context"
	"net"
	"net/http"
	"time"
	// "fmt"

	"getnoti.com/config"
	"getnoti.com/pkg/logger"
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
    logger          logger.Logger  // Add this line
}


// New creates a new HTTP server.
func New(cfg *config.Config, router *chi.Mux,logger logger.Logger) (*Server, error) {
	server := prepareHttpServer(cfg, router,logger)

    err:= server.Start()
	if err!=nil{
		return nil, err
	}
	return server,nil
}

func prepareHttpServer(cfg *config.Config, router *chi.Mux,logger logger.Logger) *Server {
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
		logger:          logger, 
	}
	return s
}

func (s *Server) Start() error {
   
    err:=s.server.ListenAndServe()
    if err!=nil{
		return err
		
	}
	return nil
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
