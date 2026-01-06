package httpUtils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"rest_api_poc/internal/shared/appError"
	"rest_api_poc/internal/shared/logger"
	"runtime/debug"
	"strings"
)

type ContextKey string

const (
	UserContextKey ContextKey = "user"
)

// UserContext is a shared minimal identity used for logging/observability.
// It is intentionally duplicated from domain context to keep httpUtils decoupled.
type UserContext struct {
	ID        string
	SessionID string
}

// Wrap adapts an error-returning handler into a standard net/http handler.
// All returned errors (and panics) are funneled through WriteError.
func Wrap(h func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// Panic is always a 500. Log stack; respond with masked message.
				logger.Error("panic recovered: %v\n%s", rec, string(debug.Stack()))
				WriteError(w, r, appError.Internal(fmt.Errorf("panic: %v", rec)))
			}
		}()

		if err := h(w, r); err != nil {
			WriteError(w, r, err)
		}
	}
}

// WriteError is the centralized error serializer + logger hook.
// Response shape is always: { "code": "...", "message": "..." }.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	ae := appError.From(err)

	userID, sessionID := extractUserContext(r)
	ip := ExtractIPAddress(r)
	logError(ae, r, userID, sessionID, ip)

	WriteJson(w, ae.HTTPStatus(), map[string]string{
		"code":    ae.ErrorCode(),
		"message": ae.PublicMessage(),
	})
}

// LogOnly logs an error with the same structured fields as WriteError, but does not write a response.
// Useful for endpoints that intentionally return a fixed status/body to avoid leaking information.
func LogOnly(r *http.Request, err error) {
	if err == nil {
		return
	}
	ae := appError.From(err)
	userID, sessionID := extractUserContext(r)
	ip := ExtractIPAddress(r)
	logError(ae, r, userID, sessionID, ip)
}

func logError(ae appError.AppError, r *http.Request, userID, sessionID, ipAddress string) {
	logMsg := "Error: %s | Method: %s | Path: %s | User: %s | Session: %s | IP: %s | Internal: %s"

	switch ae.ErrorCode() {
	case "VALIDATION_ERROR", "AUTHENTICATION_ERROR", "AUTHORIZATION_ERROR", "NOT_FOUND", "CONFLICT":
		// Expected business errors - warn level
		logger.Warn(logMsg, ae.ErrorCode(), r.Method, r.URL.Path, userID, sessionID, ipAddress, ae.InternalMessage())
	default:
		// System errors - error level
		logger.Error(logMsg, ae.ErrorCode(), r.Method, r.URL.Path, userID, sessionID, ipAddress, ae.InternalMessage())
	}
}

func extractUserContext(r *http.Request) (userID, sessionID string) {
	// Try to get from context (set by auth middleware)
	if ctx := r.Context().Value(UserContextKey); ctx != nil {
		if userCtx, ok := ctx.(*UserContext); ok {
			return userCtx.ID, userCtx.SessionID
		}
	}
	return "anonymous", "none"
}

// ExtractIPAddress extracts IP address from request, considering proxy headers.
// If behind a reverse proxy, ensure it is configured to set X-Forwarded-For / X-Real-IP correctly.
func ExtractIPAddress(r *http.Request) string {
	// X-Forwarded-For can contain multiple IPs, take the first.
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}

// WriteJSON writes JSON response
func WriteJson(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Encode directly to ResponseWriter (memory-efficient, adds newline)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// If encoding fails, send Internal Server Error
		logger.Error("Failed to encode JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WriteStatus writes only HTTP status
func WriteStatus(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

// RespondWithJSON writes JSON response (alias for WriteJson for consistency)
func RespondWithJSON(w http.ResponseWriter, statusCode int, payload any) {
	WriteJson(w, statusCode, payload)
}
