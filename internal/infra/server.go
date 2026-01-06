package infra

import (
	"context"
	"fmt"
	"net/http"
	"rest_api_poc/internal/di"
	"rest_api_poc/internal/infra/router"
	"rest_api_poc/internal/shared/logger"
	"strings"

	"github.com/fatih/color"
)

// StartServer starts the HTTP server and handles graceful shutdown
// It accepts the dependency container which manages all application dependencies
func StartServer(container *di.Container) (func(ctx context.Context) error, <-chan error) {
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

	// Run server in a goroutine so caller can handle shutdown via signals.
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	// return a function that can be called for graceful shutdown
	return func(ctx context.Context) error {
		logger.Warn("Shutting down server...")
		// Shutdown the server (caller supplies deadline/timeout via ctx).
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("Server forced to shutdown: %v", err)
			return err
		}
		logger.Info("Server stopped gracefully")
		return nil
	}, errCh
}

func printWelcome(addr string) {
	c := color.New(color.FgHiCyan)
	srvMsg := fmt.Sprintf("Starting Axil server on %s", addr)
	line := strings.Repeat("*", len(srvMsg))
	for i := 0; i < 2; i++ {
		c.Println(line)
	}
	c.Printf("**  %s  **\n", srvMsg)
	for i := 0; i < 2; i++ {
		c.Println(line)
	}
}
