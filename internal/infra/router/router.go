package router

import (
	"net/http"
	"rest_api_poc/internal/di"
	"rest_api_poc/internal/domain/auth"
	"rest_api_poc/internal/domain/health"
	"rest_api_poc/internal/domain/product"
	"rest_api_poc/internal/domain/user"

	"github.com/go-chi/chi/v5"
)

// SetupRouter configures all routes using the dependency container
// This separates routing concerns from dependency management
// The router's responsibility is to build and configure routes
func SetupRouter(container *di.Container) http.Handler {
	r := chi.NewRouter()

	// Register routes for each service module
	// Health check routes (public)
	health.RegisterRoutes(r, container.HealthHandler)

	// Auth routes (public + protected)
	auth.RegisterRoutes(r, container.AuthModule.Handler, container.AuthMiddleware, container.RoleMiddleware)

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(container.AuthMiddleware.Authenticate)

		// Product routes (read: all users, write: admin/owner only)
		product.RegisterRoutes(r, container.ProductHandler, container.RoleMiddleware)

		// User routes
		user.RegisterRoutes(r, container.UserHandler)
	})

	return r
}
