package health

import (
	"net/http"
	"rest_api_poc/internal/shared/appError"
	"rest_api_poc/internal/shared/httpUtils"
	"rest_api_poc/internal/shared/timeUtils"
	"time"
)

type Handler struct {
	service Service
}

type HealthResponse struct {
	Status    string `json:"status"`
	TimeStamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
	Database  string `json:"database"`
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

var startTime = time.Now()

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context() // Extract context from request

	// Check health including database
	healthCheck, err := h.service.CheckHealth(ctx)
	if err != nil {
		return appError.ServiceUnavailable("Service unavailable", err)
	}

	// Build complete response
	resp := HealthResponse{
		Status:    healthCheck.Status,
		TimeStamp: timeUtils.RFCTimeStampUTC(),
		Uptime:    timeUtils.Uptime(startTime),
		Database:  healthCheck.Database,
	}

	// Determine HTTP status based on database health
	statusCode := http.StatusOK
	if healthCheck.Database != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	httpUtils.WriteJson(w, statusCode, resp)
	return nil
}
