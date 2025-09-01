package log

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/mikeblum/scry.quest/conf"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		level Level
		want  slog.Level
	}{
		{LogLevelDebug, slog.LevelDebug},
		{LogLevelInfo, slog.LevelInfo},
		{LogLevelWarn, slog.LevelWarn},
		{LogLevelError, slog.LevelError},
		{"DEBUG", slog.LevelDebug},
		{"WaRn", slog.LevelWarn},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}

	for _, tt := range tests {
		got := parseLevel(tt.level)
		if got != tt.want {
			t.Errorf("parseLevel(%q) = %v, want %v", tt.level, got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	tests := []Config{
		{LogLevelDebug, LogFormatJSON, &bytes.Buffer{}},
		{LogLevelInfo, LogFormatText, &bytes.Buffer{}},
		{LogLevelWarn, LogFormatJSON, nil},
		{LogLevelError, "unknown", &bytes.Buffer{}},
	}

	for _, config := range tests {
		original := slog.Default()
		defer slog.SetDefault(original)

		New(config)

		if slog.Default() == nil {
			t.Error("New() did not set a default logger")
		}

		slog.Default().Info("test message")
	}
}

func TestNewFromEnv(t *testing.T) {
	tests := []struct {
		logLevel  string
		logFormat string
	}{
		{"", ""},
		{string(LogLevelDebug), string(LogFormatText)},
		{string(LogLevelError), string(LogFormatJSON)},
	}

	for _, tt := range tests {
		original := slog.Default()
		defer slog.SetDefault(original)

		if tt.logLevel != "" {
			_ = os.Setenv(EnvLogLevel, tt.logLevel)
			defer func() { _ = os.Unsetenv(EnvLogLevel) }()
		} else {
			_ = os.Unsetenv(EnvLogLevel)
		}

		if tt.logFormat != "" {
			_ = os.Setenv(EnvLogFormat, tt.logFormat)
			defer func() { _ = os.Unsetenv(EnvLogFormat) }()
		} else {
			_ = os.Unsetenv(EnvLogFormat)
		}

		config, err := conf.New(context.Background(), nil)
		if err != nil {
			t.Fatalf("failed to create config: %v", err)
		}

		NewFromEnv(config)

		if slog.Default() == nil {
			t.Error("NewFromEnv() did not set a default logger")
		}

		slog.Default().Info("test message from env")
	}
}

func TestNewFromEnvWithPrefixedVars(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		logFormat string
	}{
		{"default values", "", ""},
		{"debug json", "debug", "json"},
		{"error text", "error", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := slog.Default()
			defer slog.SetDefault(original)

			if tt.logLevel != "" {
				_ = os.Setenv("SCRY_LOG_LEVEL", tt.logLevel)
				defer func() { _ = os.Unsetenv("SCRY_LOG_LEVEL") }()
			} else {
				_ = os.Unsetenv("SCRY_LOG_LEVEL")
			}

			if tt.logFormat != "" {
				_ = os.Setenv("SCRY_LOG_FORMAT", tt.logFormat)
				defer func() { _ = os.Unsetenv("SCRY_LOG_FORMAT") }()
			} else {
				_ = os.Unsetenv("SCRY_LOG_FORMAT")
			}

			config, err := conf.New(context.Background(), nil)
			if err != nil {
				t.Fatalf("failed to create config: %v", err)
			}

			NewFromEnv(config)

			if slog.Default() == nil {
				t.Error("NewFromEnv() did not set a default logger")
			}

			slog.Default().Info("test message from prefixed env")
		})
	}
}

func TestConfig(t *testing.T) {
	cfg := Config{
		Level:  LogLevelDebug,
		Format: LogFormatText,
		Output: &bytes.Buffer{},
	}

	if cfg.Level != LogLevelDebug {
		t.Errorf("LOG_LEVEL = %q, want %q", cfg.Level, LogLevelDebug)
	}

	if cfg.Format != LogFormatText {
		t.Errorf("LOG_FORMAT = %q, want %q", cfg.Format, LogFormatText)
	}

	if cfg.Output == nil {
		t.Error("LOG_OUTPUT should not be nil")
	}
}

func TestLoggerOutput(t *testing.T) {
	tests := []struct {
		format Format
		level  Level
	}{
		{LogFormatJSON, LogLevelInfo},
		{LogFormatText, LogLevelDebug},
	}

	for _, tt := range tests {
		original := slog.Default()
		defer slog.SetDefault(original)

		buf := &bytes.Buffer{}

		New(Config{
			Level:  tt.level,
			Format: tt.format,
			Output: buf,
		})

		slog.Info("test message", "key", "value")

		output := buf.String()
		if output == "" {
			t.Error("expected log output, got empty string")
		}

		if tt.format == LogFormatJSON && !strings.Contains(output, `"msg":"test message"`) {
			t.Error("json format should contain structured message")
		}
		if tt.format == LogFormatText && !strings.Contains(output, "test message") {
			t.Error("text format should contain readable message")
		}
	}
}
