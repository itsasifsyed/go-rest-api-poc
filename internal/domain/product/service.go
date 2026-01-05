package product

import "context"

// Service defines the business logic interface for products
// All methods accept context for proper cancellation and timeout handling
type Service interface {
	CreateProduct(ctx context.Context, p *Product) error
	GetProduct(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context) ([]*Product, error)
	UpdateProduct(ctx context.Context, p *Product) error
	DeleteProduct(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

// NewService creates a new product service with repository dependency
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateProduct creates a new product
// Context flows from handler → service → repository for proper cancellation
func (s *service) CreateProduct(ctx context.Context, p *Product) error {
	return s.repo.CreateProduct(ctx, p)
}

// GetProduct retrieves a product by ID
// Context flows from handler → service → repository for proper cancellation
func (s *service) GetProduct(ctx context.Context, id string) (*Product, error) {
	return s.repo.GetProduct(ctx, id)
}

// ListProducts retrieves all products
// Context flows from handler → service → repository for proper cancellation
func (s *service) ListProducts(ctx context.Context) ([]*Product, error) {
	return s.repo.ListProducts(ctx)
}

// UpdateProduct updates an existing product
// Context flows from handler → service → repository for proper cancellation
func (s *service) UpdateProduct(ctx context.Context, p *Product) error {
	return s.repo.UpdateProduct(ctx, p)
}

// DeleteProduct deletes a product by ID
// Context flows from handler → service → repository for proper cancellation
func (s *service) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.DeleteProduct(ctx, id)
}
