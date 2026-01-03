package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// -------------------------
// Helper to format timestamp in UTC
// -------------------------
func timestamp() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}

type LogConfig struct {
	Type  string
	Color color.Attribute
	Block bool
	Fatal bool
}

// -------------------------
// Log functions
// -------------------------

func logger(config LogConfig, message string, args ...any) {
	c := color.New(config.Color)
	msg := fmt.Sprintf(message, args...)
	ts := timestamp()

	// print block
	if config.Block {
		// -------------------------
		// [2026-01-02 15:55:01] INFO: Loading configs
		// -------------------------
		line := strings.Repeat("-", 25)
		c.Println(line)
		c.Printf("[%s] %s: %s\n", ts, config.Type, msg)
		c.Println(line)
	} else {
		// [2026-01-02 15:55:01] INFO: Loading configs
		c.Printf("[%s] %s: %s\n", ts, config.Type, msg)
	}
	// exit for fatal logs
	if config.Fatal {
		os.Exit(1)
	}
}

// INFO - blue
func Info(message string, args ...any) {
	logger(LogConfig{Type: "INFO", Color: color.FgBlue, Block: false, Fatal: false}, message, args...)
}

func InfoBlock(message string, args ...any) {
	logger(LogConfig{Type: "INFO", Color: color.FgBlue, Block: true, Fatal: false}, message, args...)
}

// WARN - orange/yellow
func Warn(message string, args ...any) {
	logger(LogConfig{Type: "WARN", Color: color.FgYellow, Block: false, Fatal: false}, message, args...)
}

func WarnBlock(message string, args ...any) {
	logger(LogConfig{Type: "WARN", Color: color.FgYellow, Block: true, Fatal: false}, message, args...)
}

// SUCCESS - green
func Success(message string, args ...any) {
	logger(LogConfig{Type: "SUCCESS", Color: color.FgGreen, Block: false, Fatal: false}, message, args...)
}
func SuccessBlock(message string, args ...any) {
	logger(LogConfig{Type: "SUCCESS", Color: color.FgGreen, Block: true, Fatal: false}, message, args...)
}

// ERROR - red
func Error(message string, args ...any) {
	logger(LogConfig{Type: "ERROR", Color: color.FgRed, Block: false, Fatal: false}, message, args...)
}
func ErrorBlock(message string, args ...any) {
	logger(LogConfig{Type: "ERROR", Color: color.FgRed, Block: true, Fatal: false}, message, args...)
}

// FATAL - red and exit
func Fatal(message string, args ...any) {
	logger(LogConfig{Type: "FATAL", Color: color.FgRed, Block: false, Fatal: true}, message, args...)
}
func FatalBlock(message string, args ...any) {
	logger(LogConfig{Type: "FATAL", Color: color.FgRed, Block: true, Fatal: true}, message, args...)
}

// -------------------------
// Empty color blocks
// -------------------------
func logBlock(lines int, logColor color.Attribute) {
	var line = strings.Repeat("-", 25)
	c := color.New(logColor)
	for i := 0; i < lines; i++ {
		c.Printf("%s\n", line)
	}
}
func EmptyInfoBlock(lines int) {
	logBlock(lines, color.FgBlue)
}

func EmptyWarnBlock(lines int) {
	logBlock(lines, color.FgYellow)
}

func EmptySuccessBlock(lines int) {
	logBlock(lines, color.FgGreen)
}

func EmptyErrorBlock(lines int) {
	logBlock(lines, color.FgRed)
}
