package router

import (
	"net/http"
	"rest_api_poc/internal/di"
	"rest_api_poc/internal/domain/auth"
	"rest_api_poc/internal/domain/health"
	"rest_api_poc/internal/domain/product"
	"rest_api_poc/internal/domain/user"
	"rest_api_poc/internal/shared/httpUtils"

	"github.com/go-chi/chi/v5"
)

// SetupRouter configures all routes using the dependency container
// This separates routing concerns from dependency management
// The router's responsibility is to build and configure routes
func SetupRouter(container *di.Container) http.Handler {
	r := chi.NewRouter()

	// Global wrapper for error-returning handlers. Injected into domain route registration
	// to avoid package import cycles.
	wrap := func(h func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
		return httpUtils.Wrap(h)
	}

	// Register routes for each service module
	// Health check routes (public)
	health.RegisterRoutes(r, container.HealthHandler, wrap)

	// Auth routes (public + protected)
	auth.RegisterRoutes(r, container.AuthModule.Handler, container.AuthMiddleware, container.RoleMiddleware, wrap)

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(container.AuthMiddleware.Authenticate)

		// Product routes (read: all users, write: admin/owner only)
		product.RegisterRoutes(r, container.ProductHandler, container.RoleMiddleware, wrap)

		// User routes
		user.RegisterRoutes(r, container.UserHandler, wrap)
	})

	return r
}
