package main

import (
	"context"
	"rest_api_poc/config"
)

func main() {
	// create context
	ctx := context.Background()

	// Load config
	cfg := config.LoadConfig()

	// Start server, get shutdown function
	webDispose := startServer(cfg)
	defer webDispose(ctx)
}
