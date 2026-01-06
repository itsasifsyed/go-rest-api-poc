package product

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RoleMiddleware interface to avoid circular dependency
type RoleMiddleware interface {
	RequireAdmin(next http.Handler) http.Handler
}

// RegisterRoutes registers all product-related routes
// Following RESTful conventions:
//
//	GET    /v1/products      - List all products (authenticated users)
//	GET    /v1/products/{id} - Get a specific product (authenticated users)
//	POST   /v1/products      - Create a new product (admin/owner only)
//	PUT    /v1/products/{id} - Update a product (admin/owner only)
//	DELETE /v1/products/{id} - Delete a product (admin/owner only)
func RegisterRoutes(r chi.Router, h *Handler, roleMiddleware RoleMiddleware, wrap func(func(http.ResponseWriter, *http.Request) error) http.HandlerFunc) {
	r.Route("/v1/products", func(rr chi.Router) {
		// Public read access (any authenticated user)
		rr.Get("/", wrap(h.ListProducts))   // GET /v1/products - List all
		rr.Get("/{id}", wrap(h.GetProduct)) // GET /v1/products/{id} - Get one

		// Admin/Owner only routes (create, update, delete)
		rr.Group(func(rr chi.Router) {
			rr.Use(roleMiddleware.RequireAdmin)

			rr.Post("/", wrap(h.CreateProduct))       // POST /v1/products - Create
			rr.Put("/{id}", wrap(h.UpdateProduct))    // PUT /v1/products/{id} - Update
			rr.Delete("/{id}", wrap(h.DeleteProduct)) // DELETE /v1/products/{id} - Delete
		})
	})
}
