package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"rest_api_poc/internal/infra/config"
	"rest_api_poc/internal/shared/appError"
	"rest_api_poc/internal/shared/httpUtils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	config  *config.Config
}

func NewHandler(service *Service, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		config:  cfg,
	}
}

// -------------------------
// Public Endpoints
// -------------------------

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return appError.Validation("Invalid request body", err)
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		return appError.Validation("Email and password are required", nil)
	}

	// Login
	response, accessToken, refreshToken, err := h.service.Login(r.Context(), &req, r)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return appError.Authentication("Invalid email or password", err)
		}
		if errors.Is(err, ErrUserNotActive) {
			return appError.Authorization("User account is not active", err)
		}
		if errors.Is(err, ErrUserBlocked) {
			return appError.Authorization("User account has been blocked", err)
		}
		return appError.Internal(err)
	}

	// Set cookies
	h.setAccessTokenCookie(w, accessToken)
	h.setRefreshTokenCookie(w, refreshToken)

	// Also include tokens in response body for Bearer token support
	response.AccessToken = accessToken
	response.RefreshToken = refreshToken

	httpUtils.RespondWithJSON(w, http.StatusOK, response)
	return nil
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return appError.Validation("Invalid request body", err)
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return appError.Validation("All fields are required", nil)
	}

	// Register
	user, err := h.service.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return appError.Conflict("Email already exists", err)
		}
		return appError.Internal(err)
	}

	httpUtils.RespondWithJSON(w, http.StatusCreated, user)
	return nil
}

// RequestPasswordReset handles password reset request
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) error {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return appError.Validation("Invalid request body", err)
	}

	if req.Email == "" {
		return appError.Validation("Email is required", nil)
	}

	if err := h.service.RequestPasswordReset(r.Context(), req.Email); err != nil {
		return appError.Internal(err)
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset OTP has been sent",
	})
	return nil
}

// VerifyPasswordReset handles password reset verification
func (h *Handler) VerifyPasswordReset(w http.ResponseWriter, r *http.Request) error {
	var req PasswordResetVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return appError.Validation("Invalid request body", err)
	}

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		return appError.Validation("Email, OTP, and new password are required", nil)
	}

	if err := h.service.VerifyPasswordReset(r.Context(), &req); err != nil {
		if errors.Is(err, ErrInvalidOTP) {
			return appError.Validation("Invalid or expired OTP", err)
		}
		return appError.Internal(err)
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
	return nil
}

// -------------------------
// Protected Endpoints
// -------------------------

// Refresh handles token refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) error {
	// Try to get refresh token from cookie first
	refreshToken := ""
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	}

	// If no cookie, try Authorization header
	if refreshToken == "" {
		var req RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		return appError.Authentication("Refresh token is required", nil)
	}

	// Refresh tokens
	newAccessToken, newRefreshToken, err := h.service.Refresh(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, ErrExpiredToken) {
			return appError.Authentication("Refresh token has expired", err)
		}
		if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionInactive) || errors.Is(err, ErrSessionExpired) {
			return appError.Authentication("Invalid session", err)
		}
		return appError.Internal(err)
	}

	// Set new cookies
	h.setAccessTokenCookie(w, newAccessToken)
	h.setRefreshTokenCookie(w, newRefreshToken)

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":       "Tokens refreshed successfully",
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
	return nil
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	userCtx := getUserContext(r)
	if userCtx == nil {
		return appError.Authentication("Unauthorized", nil)
	}

	if err := h.service.Logout(r.Context(), userCtx.SessionID); err != nil {
		return appError.Internal(err)
	}

	// Clear cookies
	h.clearAuthCookies(w)

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
	return nil
}

// LogoutAll handles logout from all devices
func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) error {
	userCtx := getUserContext(r)
	if userCtx == nil {
		return appError.Authentication("Unauthorized", nil)
	}

	if err := h.service.LogoutAll(r.Context(), userCtx.ID); err != nil {
		return appError.Internal(err)
	}

	// Clear cookies
	h.clearAuthCookies(w)

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out from all devices successfully",
	})
	return nil
}

// GetMe returns current user information
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) error {
	userCtx := getUserContext(r)
	if userCtx == nil {
		return appError.Authentication("Unauthorized", nil)
	}

	user, err := h.service.GetMe(r.Context(), userCtx.ID)
	if err != nil {
		return appError.Internal(err)
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, user)
	return nil
}

// ChangePassword handles password change
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) error {
	userCtx := getUserContext(r)
	if userCtx == nil {
		return appError.Authentication("Unauthorized", nil)
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return appError.Validation("Invalid request body", err)
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return appError.Validation("Current password and new password are required", nil)
	}

	if err := h.service.ChangePassword(r.Context(), userCtx.ID, &req); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return appError.Authentication("Current password is incorrect", err)
		}
		return appError.Internal(err)
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
	return nil
}

// GetSessions returns all active sessions for current user
func (h *Handler) GetSessions(w http.ResponseWriter, r *http.Request) error {
	userCtx := getUserContext(r)
	if userCtx == nil {
		return appError.Authentication("Unauthorized", nil)
	}

	sessions, err := h.service.GetUserSessions(r.Context(), userCtx.ID, userCtx.SessionID)
	if err != nil {
		return appError.Internal(err)
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, sessions)
	return nil
}

// DeleteSession handles session deletion
func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	userCtx := getUserContext(r)
	if userCtx == nil {
		return appError.Authentication("Unauthorized", nil)
	}

	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		return appError.Validation("Session ID is required", nil)
	}

	if err := h.service.DeleteSession(r.Context(), sessionID, userCtx.ID); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return appError.NotFound("Session not found", err)
		}
		if err.Error() == "unauthorized to delete this session" {
			return appError.Authorization("Insufficient permissions", err)
		}
		return appError.Internal(err)
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Session deleted successfully",
	})
	return nil
}

// -------------------------
// Admin Endpoints
// -------------------------

// BlockUser handles user blocking
func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) error {
	userCtx := getUserContext(r)
	if userCtx == nil {
		return appError.Authentication("Unauthorized", nil)
	}

	targetUserID := chi.URLParam(r, "id")
	if targetUserID == "" {
		return appError.Validation("User ID is required", nil)
	}

	if err := h.service.BlockUser(r.Context(), targetUserID, userCtx.ID); err != nil {
		return appError.Internal(err)
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "User blocked successfully",
	})
	return nil
}

// UnblockUser handles user unblocking
func (h *Handler) UnblockUser(w http.ResponseWriter, r *http.Request) error {
	userCtx := getUserContext(r)
	if userCtx == nil {
		return appError.Authentication("Unauthorized", nil)
	}

	targetUserID := chi.URLParam(r, "id")
	if targetUserID == "" {
		return appError.Validation("User ID is required", nil)
	}

	if err := h.service.UnblockUser(r.Context(), targetUserID); err != nil {
		return appError.Internal(err)
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "User unblocked successfully",
	})
	return nil
}

// LogoutAllUserSessions handles admin logout of all sessions for a specific user
func (h *Handler) LogoutAllUserSessions(w http.ResponseWriter, r *http.Request) error {
	userCtx := getUserContext(r)
	if userCtx == nil {
		return appError.Authentication("Unauthorized", nil)
	}

	targetUserID := chi.URLParam(r, "id")
	if targetUserID == "" {
		return appError.Validation("User ID is required", nil)
	}

	if err := h.service.LogoutAll(r.Context(), targetUserID); err != nil {
		return appError.Internal(err)
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "All user sessions logged out successfully",
	})
	return nil
}

// -------------------------
// Helper Methods
// -------------------------

// setAccessTokenCookie sets the access token cookie
func (h *Handler) setAccessTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		HttpOnly: true,
		Secure:   h.config.WebServer.Env == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(h.config.Auth.AccessTokenLifetime.Seconds()),
	})
}

// setRefreshTokenCookie sets the refresh token cookie
func (h *Handler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		Secure:   h.config.WebServer.Env == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1/auth",
		MaxAge:   int(h.config.Auth.RefreshTokenLifetime.Seconds()),
	})
}

// clearAuthCookies clears authentication cookies
func (h *Handler) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   h.config.WebServer.Env == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   h.config.WebServer.Env == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1/auth",
		MaxAge:   -1,
	})
}

// getUserContext extracts user context from request
func getUserContext(r *http.Request) *UserContext {
	ctx := r.Context().Value(UserContextKey)
	if ctx == nil {
		return nil
	}

	userCtx, ok := ctx.(*UserContext)
	if !ok {
		return nil
	}

	return userCtx
}
