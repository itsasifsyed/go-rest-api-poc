package health

import "rest_api_poc/internal/infra/db"

// NewModule creates a new health module with all dependencies
// It follows dependency injection pattern for production-ready code
func NewModule(database db.DB) *Handler {
	svc := NewService(database)
	return NewHandler(svc)
}
