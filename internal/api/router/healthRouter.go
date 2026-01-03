package router

import (
	"net/http"
	"rest_api_poc/internal/health"
)

func RegisterHealthRoutes(mux *http.ServeMux, handler *health.Handler) {
	// v1
	mux.HandleFunc("/v1/health", handler.GetHealth)
}
