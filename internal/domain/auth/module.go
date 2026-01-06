package auth

import (
	"rest_api_poc/internal/infra/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Module encapsulates all auth dependencies
type Module struct {
	Handler    *Handler
	Service    *Service
	Repository *Repository
	JWTService *JWTService
}

// NewModule creates a new auth module with all dependencies
func NewModule(db *pgxpool.Pool, cfg *config.Config, cache AuthCache) *Module {
	// Create repository
	repo := NewRepository(db)

	// Create JWT service
	jwtService := NewJWTService(
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTIssuer,
		cfg.Auth.Audience[0],
		cfg.Auth.AccessTokenLifetime,
		cfg.Auth.RefreshTokenLifetime,
	)

	// Create service
	service := NewService(repo, cfg, cache, cfg.Cache.TTL)

	// Create handler
	handler := NewHandler(service, cfg)

	return &Module{
		Handler:    handler,
		Service:    service,
		Repository: repo,
		JWTService: jwtService,
	}
}
