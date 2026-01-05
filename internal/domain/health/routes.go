package health

import "github.com/go-chi/chi/v5"

// RegisterRoutes registers all health-related routes
// This keeps routing logic within the health domain
func RegisterRoutes(r chi.Router, handler *Handler) {
	r.Get("/health", handler.GetHealth)
}

