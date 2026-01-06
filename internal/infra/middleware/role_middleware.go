package middleware

import (
	"net/http"
	"rest_api_poc/internal/domain/auth"
	"rest_api_poc/internal/shared/appError"
	"rest_api_poc/internal/shared/httpUtils"
)

type RoleMiddleware struct{}

func NewRoleMiddleware() *RoleMiddleware {
	return &RoleMiddleware{}
}

// RequireRole creates a middleware that checks if user has required role
func (m *RoleMiddleware) RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user context
			userCtx := getUserContext(r)
			if userCtx == nil {
				httpUtils.WriteError(w, r, appError.Authentication("Unauthorized", nil))
				return
			}

			// Check if user has one of the allowed roles
			hasRole := false
			for _, role := range allowedRoles {
				if userCtx.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				httpUtils.WriteError(w, r, appError.Authorization("Insufficient permissions", nil))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a convenience middleware for admin-only routes
func (m *RoleMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireRole("owner", "admin")(next)
}

// RequireOwner is a convenience middleware for owner-only routes
func (m *RoleMiddleware) RequireOwner(next http.Handler) http.Handler {
	return m.RequireRole("owner")(next)
}

// getUserContext extracts user context from request
func getUserContext(r *http.Request) *auth.UserContext {
	ctx := r.Context().Value(auth.UserContextKey)
	if ctx == nil {
		return nil
	}

	userCtx, ok := ctx.(*auth.UserContext)
	if !ok {
		return nil
	}

	return userCtx
}
