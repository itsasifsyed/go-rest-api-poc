package infra

import (
	"context"
	"fmt"
	"net/http"
	"rest_api_poc/internal/di"
	"rest_api_poc/internal/infra/router"
	"rest_api_poc/internal/shared/logger"
	"strings"
	"time"

	"github.com/fatih/color"
)

// StartServer starts the HTTP server and handles graceful shutdown
// It accepts the dependency container which manages all application dependencies
func StartServer(container *di.Container) func(ctx context.Context) error {
	r := router.SetupRouter(container)

	addr := ":" + container.Config.WebServer.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  container.Config.WebServer.ReadTimeout,
		WriteTimeout: container.Config.WebServer.WriteTimeout,
	}
	printWelcome(addr)
	// logger.InfoBlock("Starting server on %s", addr)

	// This blocks until the server stops
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed: %v", err)
	}

	// return a function that can be called for graceful shutdown
	return func(ctx context.Context) error {
		logger.Warn("Shutting down server...")
		// Create timeout context
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		// Shutdown the server
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("Server forced to shutdown: %v", err)
			return err
		}
		logger.Success("Server stopped gracefully")
		return nil
	}
}

func printWelcome(addr string) {
	c := color.New(color.FgHiCyan)
	srvMsg := fmt.Sprintf("🚀🚀🚀 Starting Axil server on %s 🚀🚀🚀", addr)
	line := strings.Repeat("*", len(srvMsg))
	for i := 0; i < 2; i++ {
		c.Println(line)
	}
	c.Printf("**  %s  **\n", srvMsg)
	for i := 0; i < 2; i++ {
		c.Println(line)
	}
}
