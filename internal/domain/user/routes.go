package user

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all user-related routes
// Following RESTful conventions:
//
//	GET    /v1/users      - List all users
//	GET    /v1/users/{id} - Get a specific user
//	POST   /v1/users      - Create a new user
//	PUT    /v1/users/{id} - Update a user
//	DELETE /v1/users/{id} - Delete a user
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/v1/users", func(rr chi.Router) {
		rr.Get("/", h.ListUsers)         // GET /v1/users - List all
		rr.Get("/{id}", h.GetUser)       // GET /v1/users/{id} - Get one
		rr.Post("/", h.CreateUser)       // POST /v1/users - Create
		rr.Put("/{id}", h.UpdateUser)    // PUT /v1/users/{id} - Update
		rr.Delete("/{id}", h.DeleteUser) // DELETE /v1/users/{id} - Delete
	})
}
