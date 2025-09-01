// Package log provides structured logging utilities using slog
package log

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/mikeblum/scry.quest/env"
)

// Format represents the output format for logs
type Format string

// Level represents the severity level for logs
type Level string

const (
	// EnvLogFormat is the environment variable for log format
	EnvLogFormat = "LOG_FORMAT"
	// EnvLogLevel is the environment variable for log level
	EnvLogLevel = "LOG_LEVEL"
	// LogFormatJSON represents JSON log format
	LogFormatJSON Format = "json"
	// LogFormatText represents text log format
	LogFormatText Format = "text"
	// LogLevelDebug represents debug log level
	LogLevelDebug Level = "debug"
	// LogLevelInfo represents info log level
	LogLevelInfo Level = "info"
	// LogLevelWarn represents warn log level
	LogLevelWarn Level = "warn"
	// LogLevelError represents error log level
	LogLevelError Level = "error"
)

// Config holds logger configuration
type Config struct {
	Level  Level
	Format Format
	Output io.Writer
}

// New initializes the global logger with the provided configuration
func New(cfg Config) {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(cfg.Level),
		AddSource: true,
	}

	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	var handler slog.Handler
	if cfg.Format == LogFormatText {
		handler = slog.NewTextHandler(output, opts)
	} else {
		handler = slog.NewJSONHandler(output, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// NewFromEnv initializes the global logger from environment variables
func NewFromEnv() {
	New(Config{
		Level:  Level(env.GetEnv(EnvLogLevel, string(LogLevelInfo))),
		Format: Format(env.GetEnv(EnvLogFormat, string(LogFormatJSON))),
	})
}

func parseLevel(level Level) slog.Level {
	switch strings.ToLower(string(level)) {
	case string(LogLevelDebug):
		return slog.LevelDebug
	case string(LogLevelInfo):
		return slog.LevelInfo
	case string(LogLevelWarn):
		return slog.LevelWarn
	case string(LogLevelError):
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
