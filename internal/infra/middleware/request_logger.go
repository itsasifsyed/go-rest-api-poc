package middleware

import (
	"net/http"
	"rest_api_poc/internal/shared/httpUtils"
	"rest_api_poc/internal/shared/logger"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// RequestLogger logs one line per request with request_id, latency, status, and optional user context.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		userID, sessionID := func() (string, string) {
			if ctx := r.Context().Value(httpUtils.UserContextKey); ctx != nil {
				if userCtx, ok := ctx.(*httpUtils.UserContext); ok {
					return userCtx.ID, userCtx.SessionID
				}
			}
			return "anonymous", "none"
		}()

		reqID := chimw.GetReqID(r.Context())
		dur := time.Since(start)

		logger.Info(
			"request completed method=%s path=%s status=%d bytes=%d duration_ms=%d request_id=%s user_id=%s session_id=%s",
			r.Method,
			r.URL.Path,
			ww.Status(),
			ww.BytesWritten(),
			dur.Milliseconds(),
			reqID,
			userID,
			sessionID,
		)
	})
}



