package product

import (
	"encoding/json"
	"net/http"
	"rest_api_poc/internal/shared/appError"
	"rest_api_poc/internal/shared/httpUtils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return appError.Validation("Invalid request body", err)
	}

	if err := h.service.CreateProduct(ctx, &p); err != nil {
		return appError.Internal(err)
	}

	httpUtils.WriteJson(w, http.StatusCreated, p)
	return nil
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	// Extract ID from URL path using chi router
	id := chi.URLParam(r, "id")
	if id == "" {
		return appError.Validation("id parameter is required", nil)
	}

	product, err := h.service.GetProduct(ctx, id)
	if err != nil {
		if err == ErrProductNotFound {
			return appError.NotFound("Product not found", err)
		}
		return err
	}

	httpUtils.WriteJson(w, http.StatusOK, product)
	return nil
}

// ListProducts retrieves all products
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	products, err := h.service.ListProducts(ctx)
	if err != nil {
		return appError.Internal(err)
	}

	httpUtils.WriteJson(w, http.StatusOK, products)
	return nil
}

// UpdateProduct updates an existing product
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	// Extract ID from URL path
	id := chi.URLParam(r, "id")
	if id == "" {
		return appError.Validation("id parameter is required", nil)
	}

	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return appError.Validation("Invalid request body", err)
	}

	// Ensure ID from URL matches the product ID
	p.ID = id

	if err := h.service.UpdateProduct(ctx, &p); err != nil {
		if err == ErrProductNotFound {
			return appError.NotFound("Product not found", err)
		}
		return appError.Internal(err)
	}

	httpUtils.WriteJson(w, http.StatusOK, p)
	return nil
}

// DeleteProduct deletes a product by ID
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	// Extract ID from URL path
	id := chi.URLParam(r, "id")
	if id == "" {
		return appError.Validation("id parameter is required", nil)
	}

	if err := h.service.DeleteProduct(ctx, id); err != nil {
		if err == ErrProductNotFound {
			return appError.NotFound("Product not found", err)
		}
		return appError.Internal(err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
