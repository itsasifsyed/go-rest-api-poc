package auth

import (
	"context"
	"time"
)

// AuthCache is an optional performance optimization for auth checks.
// DB remains the source of truth; implementations must be safe to treat as best-effort.
type AuthCache interface {
	GetSession(ctx context.Context, sessionID string) (*CachedSession, bool, error)
	SetSession(ctx context.Context, sessionID string, s *CachedSession, ttl time.Duration) error
	DelSession(ctx context.Context, sessionID string) error

	GetUser(ctx context.Context, userID string) (*CachedUser, bool, error)
	SetUser(ctx context.Context, userID string, u *CachedUser, ttl time.Duration) error
	DelUser(ctx context.Context, userID string) error
}

type CachedSession struct {
	UserID    string    `json:"user_id"`
	IsActive  bool      `json:"is_active"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CachedUser struct {
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	IsBlocked bool   `json:"is_blocked"`
}
