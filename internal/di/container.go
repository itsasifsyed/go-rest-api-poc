package di

import (
	"rest_api_poc/internal/domain/auth"
	"rest_api_poc/internal/domain/health"
	"rest_api_poc/internal/domain/product"
	"rest_api_poc/internal/domain/user"
	"rest_api_poc/internal/infra/config"
	"rest_api_poc/internal/infra/db"
	"rest_api_poc/internal/infra/middleware"
)

// Container holds all application dependencies
// This is a simple, manual dependency injection container
// Perfect for small to medium applications (3-15 services)
// Note: Cleanup functions are handled in main.go, not here
type Container struct {
	DB             db.DB
	Config         *config.Config
	AuthModule     *auth.Module
	AuthMiddleware *middleware.AuthMiddleware
	RoleMiddleware *middleware.RoleMiddleware
	ProductHandler *product.Handler
	UserHandler    *user.Handler
	HealthHandler  *health.Handler
}

// NewContainer creates a new container with all dependencies
// This manually wires up all services - simple and explicit
func NewContainer(database db.DB, cfg *config.Config) *Container {
	// Create auth module first
	authModule := auth.NewModule(database.Pool(), cfg)

	// Create middleware with auth dependencies
	authMiddleware := middleware.NewAuthMiddleware(authModule.JWTService, authModule.Repository)
	roleMiddleware := middleware.NewRoleMiddleware()

	return &Container{
		DB:             database,
		Config:         cfg,
		AuthModule:     authModule,
		AuthMiddleware: authMiddleware,
		RoleMiddleware: roleMiddleware,
		ProductHandler: product.NewModule(database),
		UserHandler:    user.NewModule(database),
		HealthHandler:  health.NewModule(database),
	}
}
