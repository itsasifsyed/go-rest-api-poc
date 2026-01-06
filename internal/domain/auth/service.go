package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"rest_api_poc/internal/infra/config"
	"rest_api_poc/internal/shared/httpUtils"
	"rest_api_poc/internal/shared/logger"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user account is not active")
	ErrUserBlocked        = errors.New("user account has been blocked")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidOTP         = errors.New("invalid or expired OTP")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionInactive    = errors.New("session is inactive")
	ErrSessionExpired     = errors.New("session has expired")
)

type Service struct {
	repo       *Repository
	jwtService *JWTService
	config     *config.Config
	cache      AuthCache
	cacheTTL   time.Duration
}

func NewService(repo *Repository, cfg *config.Config, cache AuthCache, cacheTTL time.Duration) *Service {
	jwtService := NewJWTService(
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTIssuer,
		cfg.Auth.Audience[0],
		cfg.Auth.AccessTokenLifetime,
		cfg.Auth.RefreshTokenLifetime,
	)

	return &Service{
		repo:       repo,
		jwtService: jwtService,
		config:     cfg,
		cache:      cache,
		cacheTTL:   cacheTTL,
	}
}

// -------------------------
// Authentication Methods
// -------------------------

// Login authenticates a user and creates a new session
func (s *Service) Login(ctx context.Context, req *LoginRequest, r *http.Request) (*LoginResponse, string, string, error) {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Distinguish \"not found\" vs system failure.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", "", ErrInvalidCredentials
		}
		return nil, "", "", fmt.Errorf("get user by email: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		// Never leak account state to callers; treat as invalid credentials.
		return nil, "", "", ErrInvalidCredentials
	}

	// Check if user is blocked
	if user.IsBlocked {
		// Never leak account state to callers; treat as invalid credentials.
		return nil, "", "", ErrInvalidCredentials
	}

	// Compare password
	if err := ComparePassword(user.Password, req.Password); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	// Determine refresh token lifetime based on "stay signed in" option
	refreshLifetime := s.config.Auth.RefreshTokenLifetime
	if req.StaySignedIn {
		refreshLifetime = s.config.Auth.StaySignedInLifetime
	}

	// Create session
	_, accessToken, refreshToken, err := s.createSession(ctx, user, r, refreshLifetime)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create session: %w", err)
	}

	// Build response
	response := &LoginResponse{
		User: &UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			DeletedAt: user.DeletedAt,
		},
	}

	return response, accessToken, refreshToken, nil
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
	// Check if email already exists
	existingUser, _ := s.repo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	userID, err := s.repo.CreateUser(ctx, req.FirstName, req.LastName, req.Email, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Get created user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created user: %w", err)
	}

	logger.Info("User registered successfully: %s", user.Email)

	return &UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Refresh generates new access and refresh tokens
func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	// Hash the refresh token to look up session
	tokenHash := HashToken(refreshToken)

	// Get session
	session, err := s.repo.GetSessionByRefreshTokenHash(ctx, tokenHash)
	if err != nil {
		return "", "", ErrSessionNotFound
	}

	// Verify session is active
	if !session.IsActive {
		return "", "", ErrSessionInactive
	}

	// Verify session is not expired
	if time.Now().After(session.ExpiresAt) {
		return "", "", ErrSessionExpired
	}

	// Get user to verify they're still active
	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user: %w", err)
	}

	if !user.IsActive || user.IsBlocked {
		// Invalidate session
		_ = s.repo.InvalidateSession(ctx, session.ID)
		return "", "", ErrUserBlocked
	}

	// Generate new tokens
	newAccessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email, user.Role, session.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Calculate remaining time until session expiry
	remainingTime := time.Until(session.ExpiresAt)

	// Generate new refresh token with remaining time
	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, session.ID, remainingTime)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update session with new refresh token hash (token rotation)
	newTokenHash := HashToken(newRefreshToken)
	if err := s.repo.UpdateSessionRefreshToken(ctx, session.ID, newTokenHash); err != nil {
		return "", "", fmt.Errorf("failed to update session: %w", err)
	}

	logger.Info("Tokens refreshed for user %s (session: %s)", user.Email, session.ID)

	return newAccessToken, newRefreshToken, nil
}

// Logout invalidates the current session
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	if err := s.repo.InvalidateSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	s.cacheDelSession(ctx, sessionID)

	logger.Info("Session %s logged out", sessionID)
	return nil
}

// LogoutAll invalidates all sessions for a user
func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	sessionIDs, err := s.repo.GetActiveSessionIDsByUserID(ctx, userID)
	if err != nil {
		// Best-effort. DB remains source of truth.
		logger.Warn("failed to get active session ids for cache invalidation: %v", err)
	}
	if err := s.repo.InvalidateAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to logout all sessions: %w", err)
	}
	for _, sid := range sessionIDs {
		s.cacheDelSession(ctx, sid)
	}
	s.cacheDelUser(ctx, userID)

	logger.Info("All sessions logged out for user %s", userID)
	return nil
}

// -------------------------
// Password Management
// -------------------------

// RequestPasswordReset generates an OTP for password reset
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not
		return nil
	}

	// Generate OTP
	otp, err := GenerateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Generate secure token
	token, err := GenerateSecureToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash token
	tokenHash := HashToken(token)

	// Create password reset token
	resetToken := &PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		OTP:       otp,
		ExpiresAt: time.Now().Add(s.config.Auth.PasswordResetOTPLifetime),
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreatePasswordResetToken(ctx, resetToken); err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	// Log OTP to console (in production, send via email)
	logger.Info("===========================================")
	logger.Info("Password reset OTP for %s: %s", email, otp)
	logger.Info("OTP expires in %v", s.config.Auth.PasswordResetOTPLifetime)
	logger.Info("===========================================")

	return nil
}

// VerifyPasswordReset verifies OTP and resets password
func (s *Service) VerifyPasswordReset(ctx context.Context, req *PasswordResetVerifyRequest) error {
	// Get password reset token
	token, err := s.repo.GetPasswordResetToken(ctx, req.Email, req.OTP)
	if err != nil {
		return ErrInvalidOTP
	}
	sessionIDs, err := s.repo.GetActiveSessionIDsByUserID(ctx, token.UserID)
	if err != nil {
		logger.Warn("failed to get active session ids for cache invalidation: %v", err)
	}

	// Hash new password
	hashedPassword, err := HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	if err := s.repo.UpdateUserPassword(ctx, token.UserID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	if err := s.repo.MarkPasswordResetTokenAsUsed(ctx, token.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Invalidate all sessions for security
	if err := s.repo.InvalidateAllUserSessions(ctx, token.UserID); err != nil {
		// Best-effort cleanup; still worth surfacing upstream for centralized logging.
		return fmt.Errorf("failed to invalidate sessions after password reset: %w", err)
	}
	for _, sid := range sessionIDs {
		s.cacheDelSession(ctx, sid)
	}
	s.cacheDelUser(ctx, token.UserID)

	logger.Info("Password reset successfully for user %s", req.Email)

	return nil
}

// ChangePassword changes a user's password (requires current password)
func (s *Service) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error {
	// Get user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if err := ComparePassword(user.Password, req.CurrentPassword); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.repo.UpdateUserPassword(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	logger.Info("Password changed for user %s", user.Email)

	return nil
}

// -------------------------
// User Management
// -------------------------

// GetMe returns the current user's information
func (s *Service) GetMe(ctx context.Context, userID string) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: user.DeletedAt,
	}, nil
}

// BlockUser blocks a user and invalidates all their sessions
func (s *Service) BlockUser(ctx context.Context, userID, blockedBy string) error {
	sessionIDs, err := s.repo.GetActiveSessionIDsByUserID(ctx, userID)
	if err != nil {
		logger.Warn("failed to get active session ids for cache invalidation: %v", err)
	}
	// Block user
	if err := s.repo.BlockUser(ctx, userID, blockedBy); err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	// Invalidate all sessions
	if err := s.repo.InvalidateAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to invalidate sessions: %w", err)
	}
	for _, sid := range sessionIDs {
		s.cacheDelSession(ctx, sid)
	}
	s.cacheDelUser(ctx, userID)

	logger.Info("User %s blocked by %s", userID, blockedBy)

	return nil
}

// UnblockUser unblocks a user
func (s *Service) UnblockUser(ctx context.Context, userID string) error {
	if err := s.repo.UnblockUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}
	s.cacheDelUser(ctx, userID)

	logger.Info("User %s unblocked", userID)

	return nil
}

// -------------------------
// Session Management
// -------------------------

// GetUserSessions returns all active sessions for a user
func (s *Service) GetUserSessions(ctx context.Context, userID, currentSessionID string) ([]*SessionResponse, error) {
	sessions, err := s.repo.GetUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	var response []*SessionResponse
	for _, session := range sessions {
		response = append(response, &SessionResponse{
			ID:             session.ID,
			DeviceName:     formatDeviceName(session.DeviceInfo, session.UserAgent),
			DeviceInfo:     session.DeviceInfo,
			IPAddress:      session.IPAddress,
			LastActivityAt: session.LastActivityAt,
			ExpiresAt:      session.ExpiresAt,
			CreatedAt:      session.CreatedAt,
			IsCurrent:      session.ID == currentSessionID,
		})
	}

	return response, nil
}

// DeleteSession deletes a specific session
func (s *Service) DeleteSession(ctx context.Context, sessionID, userID string) error {
	// Verify session belongs to user
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return ErrSessionNotFound
	}

	if session.UserID != userID {
		return errors.New("unauthorized to delete this session")
	}

	if err := s.repo.InvalidateSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	s.cacheDelSession(ctx, sessionID)

	logger.Info("Session %s deleted by user %s", sessionID, userID)

	return nil
}

// -------------------------
// Helper Methods
// -------------------------

// createSession creates a new session and generates tokens
func (s *Service) createSession(ctx context.Context, user *UserWithAuth, r *http.Request, refreshLifetime time.Duration) (*Session, string, string, error) {
	// Parse device info
	deviceInfo := parseDeviceInfo(r)

	// Create session
	session := &Session{
		UserID:         user.ID,
		DeviceInfo:     deviceInfo,
		IPAddress:      httpUtils.ExtractIPAddress(r),
		UserAgent:      r.UserAgent(),
		IsActive:       true,
		LastActivityAt: time.Now(),
		ExpiresAt:      time.Now().Add(refreshLifetime),
		CreatedAt:      time.Now(),
	}

	// Generate tokens (need session ID first, so we'll use a temporary ID)
	tempSessionID := "temp"
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email, user.Role, tempSessionID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, tempSessionID, refreshLifetime)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash refresh token
	session.RefreshTokenHash = HashToken(refreshToken)

	// Create session in database
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, "", "", fmt.Errorf("failed to create session: %w", err)
	}

	// Now regenerate tokens with actual session ID
	accessToken, err = s.jwtService.GenerateAccessToken(user.ID, user.Email, user.Role, session.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = s.jwtService.GenerateRefreshToken(user.ID, session.ID, refreshLifetime)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update session with new refresh token hash
	newTokenHash := HashToken(refreshToken)
	if err := s.repo.UpdateSessionRefreshToken(ctx, session.ID, newTokenHash); err != nil {
		return nil, "", "", fmt.Errorf("failed to update session: %w", err)
	}

	// Cache session/user for faster auth checks (best-effort)
	if s.cache != nil {
		ttl := s.cacheTTL
		if until := time.Until(session.ExpiresAt); until > 0 && until < ttl {
			ttl = until
		}
		_ = s.cache.SetSession(ctx, session.ID, &CachedSession{
			UserID:    user.ID,
			IsActive:  true,
			ExpiresAt: session.ExpiresAt,
		}, ttl)
		_ = s.cache.SetUser(ctx, user.ID, &CachedUser{
			Email:     user.Email,
			Role:      user.Role,
			IsActive:  user.IsActive,
			IsBlocked: user.IsBlocked,
		}, s.cacheTTL)
	}

	return session, accessToken, refreshToken, nil
}

func (s *Service) cacheDelSession(ctx context.Context, sessionID string) {
	if s.cache == nil || sessionID == "" {
		return
	}
	if err := s.cache.DelSession(ctx, sessionID); err != nil {
		logger.Warn("auth cache del session failed: %v", err)
	}
}

func (s *Service) cacheDelUser(ctx context.Context, userID string) {
	if s.cache == nil || userID == "" {
		return
	}
	if err := s.cache.DelUser(ctx, userID); err != nil {
		logger.Warn("auth cache del user failed: %v", err)
	}
}

// parseDeviceInfo extracts device information from request
func parseDeviceInfo(r *http.Request) map[string]interface{} {
	userAgent := r.UserAgent()

	// Simple parsing (in production, use a library like mileusna/useragent)
	deviceInfo := map[string]interface{}{
		"user_agent": userAgent,
		"device":     "Unknown",
	}

	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		deviceInfo["device"] = "Mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		deviceInfo["device"] = "Tablet"
	} else if strings.Contains(ua, "postman") {
		deviceInfo["device"] = "Postman"
	} else {
		deviceInfo["device"] = "Desktop"
	}

	// Extract browser
	if strings.Contains(ua, "chrome") {
		deviceInfo["browser"] = "Chrome"
	} else if strings.Contains(ua, "firefox") {
		deviceInfo["browser"] = "Firefox"
	} else if strings.Contains(ua, "safari") {
		deviceInfo["browser"] = "Safari"
	} else if strings.Contains(ua, "edge") {
		deviceInfo["browser"] = "Edge"
	} else if strings.Contains(ua, "postman") {
		deviceInfo["browser"] = "Postman"
	} else {
		deviceInfo["browser"] = "Unknown"
	}

	return deviceInfo
}

// formatDeviceName formats a human-readable device name
func formatDeviceName(deviceInfo map[string]interface{}, userAgent string) string {
	browser, _ := deviceInfo["browser"].(string)
	device, _ := deviceInfo["device"].(string)

	if browser == "" {
		browser = "Unknown Browser"
	}
	if device == "" {
		device = "Unknown Device"
	}

	return fmt.Sprintf("%s on %s", browser, device)
}
