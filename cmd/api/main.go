package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"rest_api_poc/internal/di"
	"rest_api_poc/internal/infra"
	"rest_api_poc/internal/infra/config"
	"rest_api_poc/internal/infra/db"
	"rest_api_poc/internal/shared/logger"
	"syscall"
	"time"
)

func main() {
	// Signal-aware context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load config
	cfg := config.LoadConfig()
	logger.Init(cfg.WebServer.Env)

	// Init DB with retry mechanism and graceful shutdown
	database, dbDispose := db.SetupDB(ctx, &cfg.DB, cfg.WebServer.Env)

	// Create dependency container
	// Simple, explicit dependency injection - no magic, easy to understand
	container := di.NewContainer(database, cfg)

	// Start server (non-blocking) and wait for signal or server error
	webDispose, serverErrCh := infra.StartServer(container)

	select {
	case <-ctx.Done():
		logger.Warn("Shutdown signal received")
	case err := <-serverErrCh:
		// If the server fails to start or crashes, we still want to shutdown gracefully.
		if err != nil && err != http.ErrServerClosed {
			logger.Error("Server stopped with error: %v", err)
		} else {
			logger.Warn("Server stopped")
		}
	}

	// Use a fresh context for shutdown because ctx is canceled on signal.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown server first, then DB.
	if err := webDispose(shutdownCtx); err != nil {
		logger.Error("Server shutdown error: %v", err)
	}
	if err := dbDispose(shutdownCtx); err != nil {
		logger.Error("Database shutdown error: %v", err)
	}
}

/*
	10. CORS Middleware
	11. Combine all necessary middlewares into one
	12. Internationalization
	13. Swagger docs
	Reduce per-request DB hits in auth middleware (join query, caching, or session store).
*/
