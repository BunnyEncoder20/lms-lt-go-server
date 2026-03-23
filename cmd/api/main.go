package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-server/internal/logger"
	"go-server/internal/server"
)

func main() {
	// Initialize the server
	srv := server.NewServer(logger.NewLogger())

	// Create a context that is canceled when a termination signal is received.
	// This replaces manual signal handling and is the idiomatic way to handle OS signals.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the server in a goroutine
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for an interrupt signal (e.g., Ctrl+C or SIGTERM)
	<-ctx.Done()

	// Signal received, starting graceful shutdown
	log.Println("Interrupt signal received, shutting down gracefully...")

	// Create a context with a timeout for the shutdown process
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Graceful shutdown failed: %v", err)
	}

	log.Println("main exited")
}
