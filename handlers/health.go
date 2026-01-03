package handlers

import (
	"net/http"
	"rest_api_poc/utils"
	"time"
)

var startTime = time.Now()

type HealthResponse struct {
	Status    string `json:"status"`
	TimeStamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
}

func GetHealth(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:    "OK",
		TimeStamp: utils.RFCTimeStampUTC(),
		Uptime:    utils.Uptime(startTime),
	}
	WriteJson(w, http.StatusOK, resp)
}
