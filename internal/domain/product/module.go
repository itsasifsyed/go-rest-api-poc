package product

import "rest_api_poc/internal/infra/db"

// NewModule creates a new product module with all dependencies
// It follows dependency injection pattern for production-ready code
func NewModule(database db.DB) *Handler {
	repo := NewRepository(database)
	svc := NewService(repo)
	return NewHandler(svc)
}
