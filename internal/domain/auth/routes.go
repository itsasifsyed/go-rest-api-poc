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
func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware AuthMiddleware,
	roleMiddleware RoleMiddleware,
	wrap func(func(http.ResponseWriter, *http.Request) error) http.HandlerFunc,
) {
	// Public routes (no authentication required)
	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/login", wrap(handler.Login))
		r.Post("/register", wrap(handler.Register))
		r.Post("/reset-password", wrap(handler.RequestPasswordReset))
		r.Post("/reset-password/verify", wrap(handler.VerifyPasswordReset))

		// Protected routes (authentication required)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			r.Post("/refresh", wrap(handler.Refresh))
			r.Post("/logout", wrap(handler.Logout))
			r.Post("/logout-all", wrap(handler.LogoutAll))
			r.Get("/me", wrap(handler.GetMe))
			r.Post("/change-password", wrap(handler.ChangePassword))
			r.Get("/sessions", wrap(handler.GetSessions))
			r.Delete("/sessions/{id}", wrap(handler.DeleteSession))

			// Admin/Owner/System routes (requires admin, owner, or system role)
			r.Group(func(r chi.Router) {
				r.Use(roleMiddleware.RequireRole("owner", "admin", "system"))

				r.Post("/block-user/{id}", wrap(handler.BlockUser))
				r.Post("/unblock-user/{id}", wrap(handler.UnblockUser))
				r.Post("/logout-all-user-sessions/{id}", wrap(handler.LogoutAllUserSessions))
			})
		})
	})
}
