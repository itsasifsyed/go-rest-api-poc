package router

import (
	"net/http"
	"rest_api_poc/internal/di"
	"rest_api_poc/internal/domain/health"
	"rest_api_poc/internal/domain/product"

	"github.com/go-chi/chi/v5"
)

// SetupRouter configures all routes using the dependency container
// This separates routing concerns from dependency management
// The router's responsibility is to build and configure routes
func SetupRouter(container *di.Container) http.Handler {
	r := chi.NewRouter()

	// Register routes for each service module
	// Health check routes
	health.RegisterRoutes(r, container.HealthHandler)

	// Product routes
	product.RegisterRoutes(r, container.ProductHandler)

	// Future modules can be added here:
	// userHandler := container.UserModule()
	// user.RegisterRoutes(r, userHandler)
	//
	// authHandler := container.AuthModule()
	// auth.RegisterRoutes(r, authHandler)

	return r
}
