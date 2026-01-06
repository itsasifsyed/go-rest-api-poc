package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RoleMiddleware interface to avoid circular dependency
type RoleMiddleware interface {
	RequireAdmin(next http.Handler) http.Handler
}

// RegisterRoutes registers all user-related routes
// Following RESTful conventions:
//
//	GET    /v1/users      - List all users
//	GET    /v1/users/{id} - Get a specific user
//	POST   /v1/users      - Create a new user
//	PUT    /v1/users/{id} - Update a user
//	DELETE /v1/users/{id} - Delete a user
func RegisterRoutes(r chi.Router, h *Handler, roleMiddleware RoleMiddleware, wrap func(func(http.ResponseWriter, *http.Request) error) http.HandlerFunc) {
	r.Route("/v1/users", func(rr chi.Router) {
		// Admin/Owner only (user management)
		rr.Use(roleMiddleware.RequireAdmin)

		rr.Get("/", wrap(h.ListUsers))         // GET /v1/users - List all
		rr.Get("/{id}", wrap(h.GetUser))       // GET /v1/users/{id} - Get one
		rr.Post("/", wrap(h.CreateUser))       // POST /v1/users - Create
		rr.Put("/{id}", wrap(h.UpdateUser))    // PUT /v1/users/{id} - Update
		rr.Delete("/{id}", wrap(h.DeleteUser)) // DELETE /v1/users/{id} - Delete
	})
}
