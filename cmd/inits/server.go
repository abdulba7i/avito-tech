package inits

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func StartServer(cfg *Config, handler http.Handler) error {
	server := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.HttpServer.Timeout) * time.Second,
		WriteTimeout: time.Duration(cfg.HttpServer.Timeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.HttpServer.IdleTimeout) * time.Second,
	}

	log.Printf("Server starting on %s", cfg.HttpServer.Address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func ShutdownServer(ctx context.Context, server *http.Server) error {
	log.Println("Shutting down server...")
	return server.Shutdown(ctx)
}
