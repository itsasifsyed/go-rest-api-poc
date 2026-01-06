package appError

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is a stable machine-readable error category returned to clients.
type Code string

const (
	CodeValidation         Code = "VALIDATION_ERROR"
	CodeAuthentication     Code = "AUTHENTICATION_ERROR"
	CodeAuthorization      Code = "AUTHORIZATION_ERROR"
	CodeNotFound           Code = "NOT_FOUND"
	CodeConflict           Code = "CONFLICT"
	CodeRateLimited        Code = "RATE_LIMITED"
	CodeInternal           Code = "INTERNAL_ERROR"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
)

// AppError is the canonical application error type used for HTTP responses and logging.
//
// - PublicMessage(): what the client sees
// - InternalMessage(): what logs/observability see (never sent to client)
type AppError interface {
	error
	HTTPStatus() int
	ErrorCode() string
	PublicMessage() string
	InternalMessage() string
	Unwrap() error
}

type errImpl struct {
	code          Code
	status        int
	publicMessage string
	cause         error
}

func (e *errImpl) Error() string {
	// Error() should be safe for logs, not for clients.
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.code, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.code, e.publicMessage)
}

func (e *errImpl) HTTPStatus() int       { return e.status }
func (e *errImpl) ErrorCode() string     { return string(e.code) }
func (e *errImpl) PublicMessage() string { return e.publicMessage }
func (e *errImpl) Unwrap() error         { return e.cause }

func (e *errImpl) InternalMessage() string {
	if e.cause == nil {
		return e.publicMessage
	}
	return e.cause.Error()
}

func newErr(code Code, status int, publicMsg string, cause error) AppError {
	return &errImpl{
		code:          code,
		status:        status,
		publicMessage: publicMsg,
		cause:         cause,
	}
}

func Validation(msg string, cause error) AppError {
	return newErr(CodeValidation, http.StatusBadRequest, msg, cause)
}

func Authentication(msg string, cause error) AppError {
	return newErr(CodeAuthentication, http.StatusUnauthorized, msg, cause)
}

func Authorization(msg string, cause error) AppError {
	return newErr(CodeAuthorization, http.StatusForbidden, msg, cause)
}

func NotFound(msg string, cause error) AppError {
	return newErr(CodeNotFound, http.StatusNotFound, msg, cause)
}

func Conflict(msg string, cause error) AppError {
	return newErr(CodeConflict, http.StatusConflict, msg, cause)
}

func RateLimited(msg string, cause error) AppError {
	return newErr(CodeRateLimited, http.StatusTooManyRequests, msg, cause)
}

// Internal always masks the public message.
func Internal(cause error) AppError {
	return newErr(CodeInternal, http.StatusInternalServerError, "Something went wrong", cause)
}

func ServiceUnavailable(msg string, cause error) AppError {
	if msg == "" {
		msg = "Service unavailable"
	}
	return newErr(CodeServiceUnavailable, http.StatusServiceUnavailable, msg, cause)
}

func IsAppError(err error) (AppError, bool) {
	var ae AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
