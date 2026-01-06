package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// AuthMiddleware interface to avoid circular dependency
type AuthMiddleware interface {
	Authenticate(next http.Handler) http.Handler
}

// RoleMiddleware interface to avoid circular dependency
type RoleMiddleware interface {
	RequireAdmin(next http.Handler) http.Handler
	RequireRole(allowedRoles ...string) func(http.Handler) http.Handler
}

// RegisterRoutes registers all auth routes
func RegisterRoutes(r chi.Router, handler *Handler, authMiddleware AuthMiddleware, roleMiddleware RoleMiddleware) {
	// Public routes (no authentication required)
	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/login", handler.Login)
		r.Post("/register", handler.Register)
		r.Post("/reset-password", handler.RequestPasswordReset)
		r.Post("/reset-password/verify", handler.VerifyPasswordReset)

		// Protected routes (authentication required)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Post("/refresh", handler.Refresh)
			r.Post("/logout", handler.Logout)
			r.Post("/logout-all", handler.LogoutAll)
			r.Get("/me", handler.GetMe)
			r.Post("/change-password", handler.ChangePassword)
			r.Get("/sessions", handler.GetSessions)
			r.Delete("/sessions/{id}", handler.DeleteSession)

			// Admin/Owner/System routes (requires admin, owner, or system role)
			r.Group(func(r chi.Router) {
				r.Use(roleMiddleware.RequireRole("owner", "admin", "system"))

				r.Post("/block-user/{id}", handler.BlockUser)
				r.Post("/unblock-user/{id}", handler.UnblockUser)
				r.Post("/logout-all-user-sessions/{id}", handler.LogoutAllUserSessions)
			})
		})
	})
}
