package router

import (
	"net/http"
	"rest_api_poc/handlers"
)

func SetupRouter() http.Handler {
	mux := http.NewServeMux()

	// health check routes
	mux.HandleFunc("/health", handlers.GetHealth)

	return mux
}
