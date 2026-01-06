package middleware

import (
	"context"
	"net/http"
	"rest_api_poc/internal/domain/auth"
	"rest_api_poc/internal/shared/httpUtils"
	"rest_api_poc/internal/shared/logger"
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
			httpUtils.RespondWithError(w, http.StatusUnauthorized, "Missing authentication token")
			return
		}

		// Validate JWT
		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			if err == auth.ErrExpiredToken {
				httpUtils.RespondWithError(w, http.StatusUnauthorized, "Token has expired")
				return
			}
			logger.Warn("Invalid token: %v", err)
			httpUtils.RespondWithError(w, http.StatusUnauthorized, "Invalid authentication token")
			return
		}

		// Verify session is active
		session, err := m.repo.GetSessionByID(r.Context(), claims.SessionID)
		if err != nil {
			logger.Warn("Session not found: %s", claims.SessionID)
			httpUtils.RespondWithError(w, http.StatusUnauthorized, "Invalid session")
			return
		}

		if !session.IsActive {
			httpUtils.RespondWithError(w, http.StatusUnauthorized, "Session is inactive")
			return
		}

		// Verify user is not blocked
		user, err := m.repo.GetUserByID(r.Context(), claims.UserID)
		if err != nil {
			logger.Warn("User not found: %s", claims.UserID)
			httpUtils.RespondWithError(w, http.StatusUnauthorized, "User not found")
			return
		}

		if !user.IsActive || user.IsBlocked {
			httpUtils.RespondWithError(w, http.StatusForbidden, "User account is blocked or inactive")
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
