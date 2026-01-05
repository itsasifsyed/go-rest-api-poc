package timeUtils

import "time"

// RFCTimeStampUTC returns the current time in UTC formatted as RFC3339
func RFCTimeStampUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// TimeStampUTC returns the current time in UTC formatted as 2006-01-02 15:04:05
func TimeStampUTC() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}

// Uptime returns a human-readable duration since the given start time
func Uptime(start time.Time) string {
	return time.Since(start).String()
}
