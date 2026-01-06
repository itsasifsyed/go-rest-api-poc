package auth

import "time"

// -------------------------
// Request DTOs
// -------------------------

type LoginRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	StaySignedIn bool   `json:"stay_signed_in"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordResetVerifyRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// -------------------------
// Response DTOs
// -------------------------

type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token,omitempty"`  // Optional: for Bearer token support
	RefreshToken string        `json:"refresh_token,omitempty"` // Optional: for Bearer token support
}

type UserResponse struct {
	ID        string     `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type SessionResponse struct {
	ID             string                 `json:"id"`
	DeviceName     string                 `json:"device_name"`
	DeviceInfo     map[string]interface{} `json:"device_info,omitempty"`
	IPAddress      string                 `json:"ip_address"`
	LastActivityAt time.Time              `json:"last_activity_at"`
	ExpiresAt      time.Time              `json:"expires_at"`
	CreatedAt      time.Time              `json:"created_at"`
	IsCurrent      bool                   `json:"is_current"`
}

// -------------------------
// Domain Models
// -------------------------

type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	DeviceInfo       map[string]interface{}
	IPAddress        string
	UserAgent        string
	IsActive         bool
	LastActivityAt   time.Time
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

type PasswordResetToken struct {
	ID        string
	UserID    string
	TokenHash string
	OTP       string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type Role struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// -------------------------
// JWT Claims
// -------------------------

type AccessTokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
}

type RefreshTokenClaims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
}

// -------------------------
// Context Keys
// -------------------------

type ContextKey string

const (
	UserContextKey    ContextKey = "user"
	SessionContextKey ContextKey = "session"
)

// -------------------------
// User Context (for middleware)
// -------------------------

type UserContext struct {
	ID        string
	Email     string
	Role      string
	SessionID string
}
