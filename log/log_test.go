package log

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  slog.Level
	}{
		{LOG_LEVEL_DEBUG, slog.LevelDebug},
		{LOG_LEVEL_INFO, slog.LevelInfo},
		{LOG_LEVEL_WARN, slog.LevelWarn},
		{LOG_LEVEL_ERROR, slog.LevelError},
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
		{LOG_LEVEL_DEBUG, LOG_FORMAT_JSON, &bytes.Buffer{}},
		{LOG_LEVEL_INFO, LOG_FORMAT_TEXT, &bytes.Buffer{}},
		{LOG_LEVEL_WARN, LOG_FORMAT_JSON, nil},
		{LOG_LEVEL_ERROR, "unknown", &bytes.Buffer{}},
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
		{string(LOG_LEVEL_DEBUG), string(LOG_FORMAT_TEXT)},
		{string(LOG_LEVEL_ERROR), string(LOG_FORMAT_JSON)},
	}

	for _, tt := range tests {
		original := slog.Default()
		defer slog.SetDefault(original)

		if tt.logLevel != "" {
			_ = os.Setenv(ENV_LOG_LEVEL, tt.logLevel)
			defer func() { _ = os.Unsetenv(ENV_LOG_LEVEL) }()
		} else {
			_ = os.Unsetenv(ENV_LOG_LEVEL)
		}

		if tt.logFormat != "" {
			_ = os.Setenv(ENV_LOG_FORMAT, tt.logFormat)
			defer func() { _ = os.Unsetenv(ENV_LOG_FORMAT) }()
		} else {
			_ = os.Unsetenv(ENV_LOG_FORMAT)
		}

		NewFromEnv()

		if slog.Default() == nil {
			t.Error("NewFromEnv() did not set a default logger")
		}

		slog.Default().Info("test message from env")
	}
}

func TestConfig(t *testing.T) {
	cfg := Config{
		Level:  LOG_LEVEL_DEBUG,
		Format: LOG_FORMAT_TEXT,
		Output: &bytes.Buffer{},
	}

	if cfg.Level != LOG_LEVEL_DEBUG {
		t.Errorf("LOG_LEVEL = %q, want %q", cfg.Level, LOG_LEVEL_DEBUG)
	}

	if cfg.Format != LOG_FORMAT_TEXT {
		t.Errorf("LOG_FORMAT = %q, want %q", cfg.Format, LOG_FORMAT_TEXT)
	}

	if cfg.Output == nil {
		t.Error("LOG_OUTPUT should not be nil")
	}
}

func TestLoggerOutput(t *testing.T) {
	tests := []struct {
		format LogFormat
		level  LogLevel
	}{
		{LOG_FORMAT_JSON, LOG_LEVEL_INFO},
		{LOG_FORMAT_TEXT, LOG_LEVEL_DEBUG},
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

		if tt.format == LOG_FORMAT_JSON && !strings.Contains(output, `"msg":"test message"`) {
			t.Error("json format should contain structured message")
		}
		if tt.format == LOG_FORMAT_TEXT && !strings.Contains(output, "test message") {
			t.Error("text format should contain readable message")
		}
	}
}
