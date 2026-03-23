// Package server is responsible for handling incoming HTTP requests and routing them to the appropriate handlers. It also manages the database connection and other server-related configurations.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-server/internal/database"

	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port   int
	db     database.Service
	log    *slog.Logger
	engine *http.Server
}

func NewServer(logger *slog.Logger) *Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// this a light wrapper around the std lib's http server
	// So that it has our ports and db connection
	myServer := &Server{
		port: port,
		db:   database.New(),
		log:  logger,
	}

	// Declare Server config
	// This is the actual http engine from std http lib
	myServer.engine = &http.Server{
		Addr:         fmt.Sprintf(":%d", myServer.port),
		Handler:      myServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return myServer
}

func (s *Server) Start() error {
	s.log.Info("Starting server",
		slog.Int("port", s.port),
	)
	return s.engine.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down HTTP server...")
	if err := s.engine.Shutdown(ctx); err != nil {
		return fmt.Errorf("http shutdown error: %w", err)
	}

	s.log.Info("Closing database connections...")
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("error while closing db connections: %w", err)
	}

	s.log.Info("Server shutdown complete")
	return nil
}
