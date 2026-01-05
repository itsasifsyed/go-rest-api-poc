package health

import (
	"context"
	"rest_api_poc/internal/infra/db"
)

// Service defines the business logic interface for health checks
type Service interface {
	CheckHealth(ctx context.Context) (*HealthResponse, error)
}

type service struct {
	db db.DB
}

// NewService creates a new health service with database dependency
func NewService(database db.DB) Service {
	return &service{db: database}
}

// CheckHealth performs a comprehensive health check including database connectivity
func (s *service) CheckHealth(ctx context.Context) (*HealthResponse, error) {
	// Check database health
	dbStatus := "healthy"
	if err := s.db.Health(ctx); err != nil {
		dbStatus = "unhealthy: " + err.Error()
	}

	return &HealthResponse{
		Status:   "OK",
		Database: dbStatus,
	}, nil
}
