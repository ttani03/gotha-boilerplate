// Package main is the entry point for the application.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/ttani03/gotha-boilerplate/internal/config"
	"github.com/ttani03/gotha-boilerplate/internal/db"
	"github.com/ttani03/gotha-boilerplate/internal/server"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

const (
	defaultReadTimeout     = 10 * time.Second
	defaultWriteTimeout    = 30 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultShutdownTimeout = 10 * time.Second
)

func run() error {
	// Load .env file if present (ignored in production where env vars are set directly)
	_ = godotenv.Load()

	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		return err
	}

	// Connect to database
	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.ErrorContext(ctx, "failed to connect to database", "error", err)
		return err
	}
	defer pool.Close()

	// Create HTTP server
	handler := server.New(pool, cfg)
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		if err = srv.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(shutdownCtx, "server shutdown error", "error", err)
		}
	}()

	logger.Info("server starting", "port", cfg.Port, "env", cfg.Env)
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", "error", err)
		return err
	}

	logger.Info("server stopped")
	return nil
}
