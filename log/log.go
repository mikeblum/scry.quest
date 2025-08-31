// Package log provides structured logging utilities using slog
package log

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/mikeblum/scry.quest/env"
)

type LogFormat string
type LogLevel string

const (
	// env
	ENV_LOG_FORMAT = "LOG_FORMAT"
	ENV_LOG_LEVEL  = "LOG_LEVEL"
	// log formats
	LOG_FORMAT_JSON LogFormat = "json"
	LOG_FORMAT_TEXT LogFormat = "text"
	// log levels
	LOG_LEVEL_DEBUG LogLevel = "debug"
	LOG_LEVEL_INFO  LogLevel = "info"
	LOG_LEVEL_WARN  LogLevel = "warn"
	LOG_LEVEL_ERROR LogLevel = "error"
)

type Config struct {
	Level  LogLevel
	Format LogFormat
	Output io.Writer
}

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
	if cfg.Format == LOG_FORMAT_TEXT {
		handler = slog.NewTextHandler(output, opts)
	} else {
		handler = slog.NewJSONHandler(output, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func NewFromEnv() {
	New(Config{
		Level:  LogLevel(env.GetEnv(ENV_LOG_LEVEL, string(LOG_LEVEL_INFO))),
		Format: LogFormat(env.GetEnv(ENV_LOG_FORMAT, string(LOG_FORMAT_JSON))),
	})
}

func parseLevel(level LogLevel) slog.Level {
	switch strings.ToLower(string(level)) {
	case string(LOG_LEVEL_DEBUG):
		return slog.LevelDebug
	case string(LOG_LEVEL_INFO):
		return slog.LevelInfo
	case string(LOG_LEVEL_WARN):
		return slog.LevelWarn
	case string(LOG_LEVEL_ERROR):
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
