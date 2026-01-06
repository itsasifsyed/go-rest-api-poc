package router

import (
	"net/http"
	"rest_api_poc/internal/di"
	"rest_api_poc/internal/domain/auth"
	"rest_api_poc/internal/domain/health"
	"rest_api_poc/internal/domain/product"
	"rest_api_poc/internal/domain/user"
	"rest_api_poc/internal/infra/middleware"
	"rest_api_poc/internal/shared/httpUtils"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// SetupRouter configures all routes using the dependency container
// This separates routing concerns from dependency management
// The router's responsibility is to build and configure routes
func SetupRouter(container *di.Container) http.Handler {
	r := chi.NewRouter()

	// Standard middleware for prod readiness
	r.Use(chimw.RequestID)
	r.Use(middleware.RequestLogger)

	// CORS (config-driven). Note: wildcard origins cannot be used with credentials.
	origins := container.Config.WebServer.CORSOrigins
	allowAll := len(origins) == 0
	for _, o := range origins {
		if strings.TrimSpace(o) == "*" {
			allowAll = true
			break
		}
	}
	corsOpts := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: !allowAll,
		MaxAge:           300,
	}
	if allowAll {
		corsOpts.AllowedOrigins = []string{"*"}
	} else {
		corsOpts.AllowedOrigins = origins
	}
	r.Use(cors.Handler(corsOpts))

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
		user.RegisterRoutes(r, container.UserHandler, container.RoleMiddleware, wrap)
	})

	return r
}
