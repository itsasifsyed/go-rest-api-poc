package user

import (
	"context"
	"errors"
	"rest_api_poc/internal/infra/db"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
)

type Repository interface {
	CreateUser(ctx context.Context, u *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	ListUsers(ctx context.Context) ([]*User, error)
	UpdateUser(ctx context.Context, u *User) error
	DeleteUser(ctx context.Context, id string) error
}

type repository struct {
	db db.DB
}

// NewRepository creates a new user repository with database dependency
func NewRepository(database db.DB) Repository {
	return &repository{db: database}
}

func (r *repository) CreateUser(ctx context.Context, u *User) error {
	_, err := r.db.Pool().Exec(ctx,
		"INSERT INTO users (id, first_name, last_name, email) VALUES ($1, $2, $3, $4)",
		u.ID, u.FirstName, u.LastName, u.Email,
	)
	return err
}

func (r *repository) GetUser(ctx context.Context, id string) (*User, error) {
	row := r.db.Pool().QueryRow(ctx,
		"SELECT id, first_name, last_name, email FROM users WHERE id=$1", id,
	)
	user := &User{}
	if err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email); err != nil {
		return nil, err
	}
	return user, nil
}

// ListUsers retrieves all users
func (r *repository) ListUsers(ctx context.Context) ([]*User, error) {
	rows, err := r.db.Pool().Query(ctx,
		"SELECT id, first_name, last_name, email FROM users ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser updates an existing user
func (r *repository) UpdateUser(ctx context.Context, u *User) error {
	result, err := r.db.Pool().Exec(ctx,
		"UPDATE users SET first_name=$1, last_name=$2, email=$3 WHERE id=$4",
		u.FirstName, u.LastName, u.Email, u.ID,
	)
	if err != nil {
		return err
	}

	// Check if any row was actually updated
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser deletes a user by ID
func (r *repository) DeleteUser(ctx context.Context, id string) error {
	result, err := r.db.Pool().Exec(ctx,
		"DELETE FROM users WHERE id=$1", id,
	)
	if err != nil {
		return err
	}

	// Check if any row was actually deleted
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
