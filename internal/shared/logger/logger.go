package logger

import (
	"fmt"
	"log/slog"
	"os"
	"rest_api_poc/internal/shared/timeUtils"
	"strings"
)

// Init configures the global logger. Call once from main after config is loaded.
// env: "dev" enables human-readable logs; everything else uses JSON.
func Init(env string) {
	var h slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if strings.EqualFold(env, "dev") || strings.EqualFold(env, "local") {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	// Add stable fields every log line.
	base := slog.New(h).With(
		slog.String("ts", timeUtils.TimeStampUTC()),
		slog.String("service", "rest_api_poc"),
	)
	slog.SetDefault(base)
}

func Info(message string, args ...any)  { slog.Info(fmt.Sprintf(message, args...)) }
func Warn(message string, args ...any)  { slog.Warn(fmt.Sprintf(message, args...)) }
func Error(message string, args ...any) { slog.Error(fmt.Sprintf(message, args...)) }

func Fatal(message string, args ...any) {
	slog.Error(fmt.Sprintf(message, args...))
	os.Exit(1)
}
