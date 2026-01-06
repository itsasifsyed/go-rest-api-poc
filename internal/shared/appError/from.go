package appError

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// From normalizes arbitrary errors into an AppError.
//
// Keep this free of domain package imports to avoid import cycles; domain handlers/services
// should translate domain-specific errors into AppError explicitly.
func From(err error) AppError {
	if err == nil {
		return nil
	}
	if ae, ok := IsAppError(err); ok {
		return ae
	}

	// Request context / timeouts
	if errors.Is(err, context.Canceled) {
		// Closest standard code (Go stdlib constant). Some proxies use 499, but net/http has no constant.
		return newErr(CodeServiceUnavailable, http.StatusRequestTimeout, "Request canceled", err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return newErr(CodeServiceUnavailable, http.StatusGatewayTimeout, "Request timed out", err)
	}

	// Common DB not-found mapping
	if errors.Is(err, pgx.ErrNoRows) {
		return NotFound("Not found", err)
	}

	return Internal(err)
}
