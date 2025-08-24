// Package log provides structured logging utilities using slog
package log

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/mikeblum/scry.quest/env"
)

// Config holds the logger configuration
type Config struct {
	Level  string `env:"LOG_LEVEL" env-default:"info"`
	Format string `env:"LOG_FORMAT" env-default:"json"`
	Output io.Writer
}

// New creates and sets a new default logger with the given configuration
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
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(output, opts)
	} else {
		handler = slog.NewJSONHandler(output, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// NewFromEnv creates a logger configured from environment variables
func NewFromEnv() {
	New(Config{
		Level:  env.GetEnv("LOG_LEVEL", "info"),
		Format: env.GetEnv("LOG_FORMAT", "json"),
	})
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
