package product

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Extract context from request

	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateProduct(ctx, &p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Extract context from request

	// Extract ID from URL path using chi router
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProduct(ctx, id)
	if err != nil {
		if err == ErrProductNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// ListProducts retrieves all products
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Extract context from request

	products, err := h.service.ListProducts(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// UpdateProduct updates an existing product
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Extract context from request

	// Extract ID from URL path
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure ID from URL matches the product ID
	p.ID = id

	if err := h.service.UpdateProduct(ctx, &p); err != nil {
		if err == ErrProductNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// DeleteProduct deletes a product by ID
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Extract context from request

	// Extract ID from URL path
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteProduct(ctx, id); err != nil {
		if err == ErrProductNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
