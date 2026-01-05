package product

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all product-related routes
// Following RESTful conventions:
//
//	GET    /v1/products      - List all products
//	GET    /v1/products/{id} - Get a specific product
//	POST   /v1/products      - Create a new product
//	PUT    /v1/products/{id} - Update a product
//	DELETE /v1/products/{id} - Delete a product
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/v1/products", func(rr chi.Router) {
		rr.Get("/", h.ListProducts)         // GET /v1/products - List all
		rr.Get("/{id}", h.GetProduct)       // GET /v1/products/{id} - Get one
		rr.Post("/", h.CreateProduct)       // POST /v1/products - Create
		rr.Put("/{id}", h.UpdateProduct)    // PUT /v1/products/{id} - Update
		rr.Delete("/{id}", h.DeleteProduct) // DELETE /v1/products/{id} - Delete
	})
}
