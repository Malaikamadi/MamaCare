package http

import (
	"context"
	"net/http"
	"time"

	"github.com/mamacare/services/pkg/logger"
)

// Config contains HTTP server configuration
type Config struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// Server represents an HTTP server
type Server struct {
	server  *http.Server
	logger  logger.Logger
	config  Config
	handler http.Handler
}

// NewServer creates a new HTTP server
func NewServer(config Config, handler http.Handler, logger logger.Logger) *Server {
	return &Server{
		config:  config,
		handler: handler,
		logger:  logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         s.config.Address,
		Handler:      s.handler,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	s.logger.Info("Starting HTTP server", logger.Field{Key: "address", Value: s.config.Address})
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()
	
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("HTTP server shutdown error", err)
		return err
	}
	
	s.logger.Info("HTTP server stopped")
	return nil
}