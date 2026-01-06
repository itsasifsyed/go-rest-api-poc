package user

import (
	"encoding/json"
	"net/http"
	"rest_api_poc/internal/shared/appError"
	"rest_api_poc/internal/shared/httpUtils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return appError.Validation("Invalid request body", err)
	}

	if err := h.service.CreateUser(ctx, &u); err != nil {
		return appError.Internal(err)
	}

	httpUtils.WriteJson(w, http.StatusCreated, u)
	return nil
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	// Extract ID from URL path using chi router
	id := chi.URLParam(r, "id")
	if id == "" {
		return appError.Validation("id parameter is required", nil)
	}

	user, err := h.service.GetUser(ctx, id)
	if err != nil {
		if err == ErrUserNotFound {
			return appError.NotFound("User not found", err)
		}
		// pgx.ErrNoRows and other common errors are normalized centrally
		return err
	}

	httpUtils.WriteJson(w, http.StatusOK, user)
	return nil
}

// ListUsers retrieves all users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	users, err := h.service.ListUsers(ctx)
	if err != nil {
		return appError.Internal(err)
	}

	httpUtils.WriteJson(w, http.StatusOK, users)
	return nil
}

// UpdateUser updates an existing user
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	// Extract ID from URL path
	id := chi.URLParam(r, "id")
	if id == "" {
		return appError.Validation("id parameter is required", nil)
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return appError.Validation("Invalid request body", err)
	}

	// Ensure ID from URL matches the user ID
	u.ID = id

	if err := h.service.UpdateUser(ctx, &u); err != nil {
		if err == ErrUserNotFound {
			return appError.NotFound("User not found", err)
		}
		return appError.Internal(err)
	}

	httpUtils.WriteJson(w, http.StatusOK, u)
	return nil
}

// DeleteUser deletes a user by ID
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	// Extract ID from URL path
	id := chi.URLParam(r, "id")
	if id == "" {
		return appError.Validation("id parameter is required", nil)
	}

	if err := h.service.DeleteUser(ctx, id); err != nil {
		if err == ErrUserNotFound {
			return appError.NotFound("User not found", err)
		}
		return appError.Internal(err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
