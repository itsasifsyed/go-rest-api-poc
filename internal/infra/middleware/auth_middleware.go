package middleware

import (
	"context"
	"net/http"
	"rest_api_poc/internal/domain/auth"
	"rest_api_poc/internal/shared/appError"
	"rest_api_poc/internal/shared/httpUtils"
	"strings"
)

type AuthMiddleware struct {
	jwtService *auth.JWTService
	repo       *auth.Repository
}

func NewAuthMiddleware(jwtService *auth.JWTService, repo *auth.Repository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		repo:       repo,
	}
}

// Authenticate validates JWT and attaches user context to request
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from cookie or Authorization header
		token := m.extractToken(r)
		if token == "" {
			httpUtils.WriteError(w, r, appError.Authentication("Missing authentication token", nil))
			return
		}

		// Validate JWT
		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			if err == auth.ErrExpiredToken {
				httpUtils.WriteError(w, r, appError.Authentication("Token has expired", err))
				return
			}
			httpUtils.WriteError(w, r, appError.Authentication("Invalid authentication token", err))
			return
		}

		// Verify session is active
		session, err := m.repo.GetSessionByID(r.Context(), claims.SessionID)
		if err != nil {
			httpUtils.WriteError(w, r, appError.Authentication("Invalid session", err))
			return
		}

		if !session.IsActive {
			httpUtils.WriteError(w, r, appError.Authentication("Invalid session", nil))
			return
		}

		// Verify user is not blocked
		user, err := m.repo.GetUserByID(r.Context(), claims.UserID)
		if err != nil {
			// Avoid user enumeration; treat as invalid auth.
			httpUtils.WriteError(w, r, appError.Authentication("Invalid authentication token", err))
			return
		}

		if !user.IsActive || user.IsBlocked {
			httpUtils.WriteError(w, r, appError.Authorization("User account is blocked or inactive", nil))
			return
		}

		// Attach user context to request
		userCtx := &auth.UserContext{
			ID:        claims.UserID,
			Email:     claims.Email,
			Role:      claims.Role,
			SessionID: claims.SessionID,
		}

		ctx := context.WithValue(r.Context(), auth.UserContextKey, userCtx)
		// Also set a shared minimal user context for httpUtils logging/extraction (decoupled from domain packages).
		ctx = context.WithValue(ctx, httpUtils.UserContextKey, &httpUtils.UserContext{
			ID:        userCtx.ID,
			SessionID: userCtx.SessionID,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractToken extracts JWT from cookie or Authorization header
func (m *AuthMiddleware) extractToken(r *http.Request) string {
	// Try cookie first (preferred for browsers)
	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Try Authorization header (for API clients like Postman)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Bearer token format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	return ""
}
