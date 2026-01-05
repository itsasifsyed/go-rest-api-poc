package health

import (
	"net/http"
	"rest_api_poc/internal/shared/httpUtils"
	"rest_api_poc/internal/shared/timeUtils"
	"time"
)

type Handler struct{}

type HealthResponse struct {
	Status    string `json:"status"`
	TimeStamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
}

func NewHandler() *Handler {
	return &Handler{}
}

var startTime = time.Now()

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:    "OK",
		TimeStamp: timeUtils.RFCTimeStampUTC(),
		Uptime:    timeUtils.Uptime(startTime),
	}
	httpUtils.WriteJson(w, http.StatusOK, resp)
}
