package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"rest_api_poc/internal/infra/config"
	"rest_api_poc/internal/shared/httpUtils"
	"rest_api_poc/internal/shared/logger"

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
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Login
	response, accessToken, refreshToken, err := h.service.Login(r.Context(), &req, r)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpUtils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		if errors.Is(err, ErrUserNotActive) {
			httpUtils.RespondWithError(w, http.StatusForbidden, "User account is not active")
			return
		}
		if errors.Is(err, ErrUserBlocked) {
			httpUtils.RespondWithError(w, http.StatusForbidden, "User account has been blocked")
			return
		}
		logger.Error("Login error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to login")
		return
	}

	// Set cookies
	h.setAccessTokenCookie(w, accessToken)
	h.setRefreshTokenCookie(w, refreshToken)

	// Also include tokens in response body for Bearer token support
	response.AccessToken = accessToken
	response.RefreshToken = refreshToken

	httpUtils.RespondWithJSON(w, http.StatusOK, response)
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "All fields are required")
		return
	}

	// Register
	user, err := h.service.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			httpUtils.RespondWithError(w, http.StatusConflict, "Email already exists")
			return
		}
		logger.Error("Registration error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to register")
		return
	}

	httpUtils.RespondWithJSON(w, http.StatusCreated, user)
}

// RequestPasswordReset handles password reset request
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	if err := h.service.RequestPasswordReset(r.Context(), req.Email); err != nil {
		logger.Error("Password reset request error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to process password reset request")
		return
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset OTP has been sent",
	})
}

// VerifyPasswordReset handles password reset verification
func (h *Handler) VerifyPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Email, OTP, and new password are required")
		return
	}

	if err := h.service.VerifyPasswordReset(r.Context(), &req); err != nil {
		if errors.Is(err, ErrInvalidOTP) {
			httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid or expired OTP")
			return
		}
		logger.Error("Password reset verification error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}

// -------------------------
// Protected Endpoints
// -------------------------

// Refresh handles token refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
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
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Refresh token is required")
		return
	}

	// Refresh tokens
	newAccessToken, newRefreshToken, err := h.service.Refresh(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, ErrExpiredToken) {
			httpUtils.RespondWithError(w, http.StatusUnauthorized, "Refresh token has expired")
			return
		}
		if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionInactive) {
			httpUtils.RespondWithError(w, http.StatusUnauthorized, "Invalid session")
			return
		}
		logger.Error("Token refresh error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to refresh token")
		return
	}

	// Set new cookies
	h.setAccessTokenCookie(w, newAccessToken)
	h.setRefreshTokenCookie(w, newRefreshToken)

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message":       "Tokens refreshed successfully",
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userCtx := getUserContext(r)
	if userCtx == nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.service.Logout(r.Context(), userCtx.SessionID); err != nil {
		logger.Error("Logout error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	// Clear cookies
	h.clearAuthCookies(w)

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// LogoutAll handles logout from all devices
func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userCtx := getUserContext(r)
	if userCtx == nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.service.LogoutAll(r.Context(), userCtx.ID); err != nil {
		logger.Error("Logout all error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to logout from all devices")
		return
	}

	// Clear cookies
	h.clearAuthCookies(w)

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out from all devices successfully",
	})
}

// GetMe returns current user information
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userCtx := getUserContext(r)
	if userCtx == nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.service.GetMe(r.Context(), userCtx.ID)
	if err != nil {
		logger.Error("Get me error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user information")
		return
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, user)
}

// ChangePassword handles password change
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userCtx := getUserContext(r)
	if userCtx == nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Current password and new password are required")
		return
	}

	if err := h.service.ChangePassword(r.Context(), userCtx.ID, &req); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpUtils.RespondWithError(w, http.StatusUnauthorized, "Current password is incorrect")
			return
		}
		logger.Error("Change password error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to change password")
		return
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// GetSessions returns all active sessions for current user
func (h *Handler) GetSessions(w http.ResponseWriter, r *http.Request) {
	userCtx := getUserContext(r)
	if userCtx == nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessions, err := h.service.GetUserSessions(r.Context(), userCtx.ID, userCtx.SessionID)
	if err != nil {
		logger.Error("Get sessions error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to get sessions")
		return
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, sessions)
}

// DeleteSession handles session deletion
func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	userCtx := getUserContext(r)
	if userCtx == nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	if err := h.service.DeleteSession(r.Context(), sessionID, userCtx.ID); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			httpUtils.RespondWithError(w, http.StatusNotFound, "Session not found")
			return
		}
		logger.Error("Delete session error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete session")
		return
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Session deleted successfully",
	})
}

// -------------------------
// Admin Endpoints
// -------------------------

// BlockUser handles user blocking
func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	userCtx := getUserContext(r)
	if userCtx == nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	targetUserID := chi.URLParam(r, "id")
	if targetUserID == "" {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	if err := h.service.BlockUser(r.Context(), targetUserID, userCtx.ID); err != nil {
		logger.Error("Block user error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to block user")
		return
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "User blocked successfully",
	})
}

// UnblockUser handles user unblocking
func (h *Handler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	userCtx := getUserContext(r)
	if userCtx == nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	targetUserID := chi.URLParam(r, "id")
	if targetUserID == "" {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	if err := h.service.UnblockUser(r.Context(), targetUserID); err != nil {
		logger.Error("Unblock user error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to unblock user")
		return
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "User unblocked successfully",
	})
}

// LogoutAllUserSessions handles admin logout of all sessions for a specific user
func (h *Handler) LogoutAllUserSessions(w http.ResponseWriter, r *http.Request) {
	userCtx := getUserContext(r)
	if userCtx == nil {
		httpUtils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	targetUserID := chi.URLParam(r, "id")
	if targetUserID == "" {
		httpUtils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	if err := h.service.LogoutAll(r.Context(), targetUserID); err != nil {
		logger.Error("Admin logout all sessions error: %v", err)
		httpUtils.RespondWithError(w, http.StatusInternalServerError, "Failed to logout all user sessions")
		return
	}

	httpUtils.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "All user sessions logged out successfully",
	})
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
