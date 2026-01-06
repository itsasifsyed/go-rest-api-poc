package health

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all health-related routes
// This keeps routing logic within the health domain
func RegisterRoutes(r chi.Router, handler *Handler, wrap func(func(http.ResponseWriter, *http.Request) error) http.HandlerFunc) {
	r.Get("/health", wrap(handler.GetHealth))
}

