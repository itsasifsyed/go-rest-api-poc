package user

import "context"

// Service defines the business logic interface for users
// All methods accept context for proper cancellation and timeout handling
type Service interface {
	CreateUser(ctx context.Context, u *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	ListUsers(ctx context.Context) ([]*User, error)
	UpdateUser(ctx context.Context, u *User) error
	DeleteUser(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

// NewService creates a new user service with repository dependency
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateUser creates a new user
// Context flows from handler → service → repository for proper cancellation
func (s *service) CreateUser(ctx context.Context, u *User) error {
	return s.repo.CreateUser(ctx, u)
}

// GetUser retrieves a user by ID
// Context flows from handler → service → repository for proper cancellation
func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
	return s.repo.GetUser(ctx, id)
}

// ListUsers retrieves all users
// Context flows from handler → service → repository for proper cancellation
func (s *service) ListUsers(ctx context.Context) ([]*User, error) {
	return s.repo.ListUsers(ctx)
}

// UpdateUser updates an existing user
// Context flows from handler → service → repository for proper cancellation
func (s *service) UpdateUser(ctx context.Context, u *User) error {
	return s.repo.UpdateUser(ctx, u)
}

// DeleteUser deletes a user by ID
// Context flows from handler → service → repository for proper cancellation
func (s *service) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}
