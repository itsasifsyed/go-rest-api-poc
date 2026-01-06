package middleware

import (
	"context"
	"net/http"
	"rest_api_poc/internal/domain/auth"
	"rest_api_poc/internal/infra/config"
	"rest_api_poc/internal/shared/appError"
	"rest_api_poc/internal/shared/httpUtils"
	"rest_api_poc/internal/shared/logger"
	"strings"
	"time"
)

type AuthMiddleware struct {
	jwtService *auth.JWTService
	repo       *auth.Repository
	cache      auth.AuthCache
	cacheTTL   time.Duration
}

func NewAuthMiddleware(jwtService *auth.JWTService, repo *auth.Repository, cache auth.AuthCache, cfg *config.Config) *AuthMiddleware {
	ttl := time.Hour
	if cfg != nil && cfg.Cache.TTL > 0 {
		ttl = cfg.Cache.TTL
	}
	return &AuthMiddleware{
		jwtService: jwtService,
		repo:       repo,
		cache:      cache,
		cacheTTL:   ttl,
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

		now := time.Now()

		// Verify session is active
		var sessionUserID string
		var sessionExpiresAt time.Time
		if m.cache != nil {
			if cs, ok, err := m.cache.GetSession(r.Context(), claims.SessionID); err != nil {
				logger.Warn("auth cache session get failed: %v", err)
			} else if ok && cs != nil {
				if !cs.IsActive || (!cs.ExpiresAt.IsZero() && now.After(cs.ExpiresAt)) {
					// Best-effort cleanup
					_ = m.cache.DelSession(r.Context(), claims.SessionID)
					httpUtils.WriteError(w, r, appError.Authentication("Invalid session", nil))
					return
				}
				sessionUserID = cs.UserID
				sessionExpiresAt = cs.ExpiresAt
			}
		}
		if sessionUserID == "" {
			session, err := m.repo.GetSessionByID(r.Context(), claims.SessionID)
			if err != nil {
				httpUtils.WriteError(w, r, appError.Authentication("Invalid session", err))
				return
			}
			if !session.IsActive || now.After(session.ExpiresAt) {
				httpUtils.WriteError(w, r, appError.Authentication("Invalid session", nil))
				return
			}
			sessionUserID = session.UserID
			sessionExpiresAt = session.ExpiresAt

			// Populate cache (best-effort)
			if m.cache != nil {
				ttl := m.cacheTTL
				if !sessionExpiresAt.IsZero() {
					if until := time.Until(sessionExpiresAt); until > 0 && until < ttl {
						ttl = until
					}
				}
				_ = m.cache.SetSession(r.Context(), claims.SessionID, &auth.CachedSession{
					UserID:    session.UserID,
					IsActive:  session.IsActive,
					ExpiresAt: session.ExpiresAt,
				}, ttl)
			}
		}

		// Verify user is not blocked
		// Session and token must agree on user id.
		if sessionUserID != "" && sessionUserID != claims.UserID {
			httpUtils.WriteError(w, r, appError.Authentication("Invalid authentication token", nil))
			return
		}

		var userEmail, userRole string
		var userIsActive, userIsBlocked bool
		foundUser := false
		if m.cache != nil {
			if cu, ok, err := m.cache.GetUser(r.Context(), claims.UserID); err != nil {
				logger.Warn("auth cache user get failed: %v", err)
			} else if ok && cu != nil {
				userEmail = cu.Email
				userRole = cu.Role
				userIsActive = cu.IsActive
				userIsBlocked = cu.IsBlocked
				foundUser = true
			}
		}
		if !foundUser {
			user, err := m.repo.GetUserByID(r.Context(), claims.UserID)
			if err != nil {
				// Avoid user enumeration; treat as invalid auth.
				httpUtils.WriteError(w, r, appError.Authentication("Invalid authentication token", err))
				return
			}
			userEmail = user.Email
			userRole = user.Role
			userIsActive = user.IsActive
			userIsBlocked = user.IsBlocked
			foundUser = true

			// Populate cache (best-effort)
			if m.cache != nil {
				_ = m.cache.SetUser(r.Context(), claims.UserID, &auth.CachedUser{
					Email:     user.Email,
					Role:      user.Role,
					IsActive:  user.IsActive,
					IsBlocked: user.IsBlocked,
				}, m.cacheTTL)
			}
		}
		if !userIsActive || userIsBlocked {
			httpUtils.WriteError(w, r, appError.Authorization("User account is blocked or inactive", nil))
			return
		}

		// Attach user context to request
		userCtx := &auth.UserContext{
			ID:        claims.UserID,
			Email:     userEmail,
			Role:      userRole,
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
