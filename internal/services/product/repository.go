package product

import (
	"context"
	"errors"
	"rest_api_poc/internal/platform/db"
)

var (
	// ErrProductNotFound is returned when a product is not found
	ErrProductNotFound = errors.New("product not found")
)

type Repository interface {
	CreateProduct(ctx context.Context, p *Product) error
	GetProduct(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context) ([]*Product, error)
	UpdateProduct(ctx context.Context, p *Product) error
	DeleteProduct(ctx context.Context, id string) error
}

type repository struct {
	db db.DB
}

// NewRepository creates a new product repository with database dependency
func NewRepository(database db.DB) Repository {
	return &repository{db: database}
}

func (r *repository) CreateProduct(ctx context.Context, p *Product) error {
	_, err := r.db.Pool().Exec(ctx,
		"INSERT INTO products (id, name, price) VALUES ($1, $2, $3)",
		p.ID, p.Name, p.Price,
	)
	return err
}

func (r *repository) GetProduct(ctx context.Context, id string) (*Product, error) {
	row := r.db.Pool().QueryRow(ctx,
		"SELECT id, name, price FROM products WHERE id=$1", id,
	)
	prod := &Product{}
	if err := row.Scan(&prod.ID, &prod.Name, &prod.Price); err != nil {
		return nil, err
	}
	return prod, nil
}

// ListProducts retrieves all products
func (r *repository) ListProducts(ctx context.Context) ([]*Product, error) {
	rows, err := r.db.Pool().Query(ctx,
		"SELECT id, name, price FROM products ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		prod := &Product{}
		if err := rows.Scan(&prod.ID, &prod.Name, &prod.Price); err != nil {
			return nil, err
		}
		products = append(products, prod)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// UpdateProduct updates an existing product
func (r *repository) UpdateProduct(ctx context.Context, p *Product) error {
	result, err := r.db.Pool().Exec(ctx,
		"UPDATE products SET name=$1, price=$2 WHERE id=$3",
		p.Name, p.Price, p.ID,
	)
	if err != nil {
		return err
	}

	// Check if any row was actually updated
	if result.RowsAffected() == 0 {
		return ErrProductNotFound
	}

	return nil
}

// DeleteProduct deletes a product by ID
func (r *repository) DeleteProduct(ctx context.Context, id string) error {
	result, err := r.db.Pool().Exec(ctx,
		"DELETE FROM products WHERE id=$1", id,
	)
	if err != nil {
		return err
	}

	// Check if any row was actually deleted
	if result.RowsAffected() == 0 {
		return ErrProductNotFound
	}

	return nil
}
