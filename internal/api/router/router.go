package router

import (
	"net/http"
	"rest_api_poc/internal/health"
)

func SetupRouter() http.Handler {
	mux := http.NewServeMux()

	// health check routes
	healthHandler := health.NewHandler()
	RegisterHealthRoutes(mux, healthHandler)
	return mux
}
