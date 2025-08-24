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
		name  string
		level string
		want  slog.Level
	}{
		{
			name:  "debug level",
			level: "debug",
			want:  slog.LevelDebug,
		},
		{
			name:  "info level",
			level: "info",
			want:  slog.LevelInfo,
		},
		{
			name:  "warn level",
			level: "warn",
			want:  slog.LevelWarn,
		},
		{
			name:  "error level",
			level: "error",
			want:  slog.LevelError,
		},
		{
			name:  "uppercase level",
			level: "DEBUG",
			want:  slog.LevelDebug,
		},
		{
			name:  "mixed case level",
			level: "WaRn",
			want:  slog.LevelWarn,
		},
		{
			name:  "unknown level defaults to info",
			level: "unknown",
			want:  slog.LevelInfo,
		},
		{
			name:  "empty level defaults to info",
			level: "",
			want:  slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLevel(tt.level)
			if got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "json handler with debug level",
			config: Config{
				Level:  "debug",
				Format: "json",
				Output: &bytes.Buffer{},
			},
		},
		{
			name: "text handler with info level",
			config: Config{
				Level:  "info",
				Format: "text",
				Output: &bytes.Buffer{},
			},
		},
		{
			name: "default output when nil",
			config: Config{
				Level:  "warn",
				Format: "json",
				Output: nil,
			},
		},
		{
			name: "unknown format defaults to json",
			config: Config{
				Level:  "error",
				Format: "unknown",
				Output: &bytes.Buffer{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := slog.Default()
			defer slog.SetDefault(original)

			New(tt.config)

			logger := slog.Default()
			if logger == nil {
				t.Error("New() did not set a default logger")
			}

			logger.Info("test message")
		})
	}
}

func TestNewFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		logFormat string
	}{
		{
			name:      "default values when env vars not set",
			logLevel:  "",
			logFormat: "",
		},
		{
			name:      "custom level and format from env",
			logLevel:  "debug",
			logFormat: "text",
		},
		{
			name:      "mixed env values",
			logLevel:  "error",
			logFormat: "json",
		},
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

			NewFromEnv()

			logger := slog.Default()
			if logger == nil {
				t.Error("NewFromEnv() did not set a default logger")
			}

			logger.Info("test message from env")
		})
	}
}

func TestConfig(t *testing.T) {
	const textFormat = "text"
	cfg := Config{
		Level:  "debug",
		Format: textFormat,
		Output: &bytes.Buffer{},
	}

	if cfg.Level != "debug" {
		t.Errorf("Config.Level = %q, want %q", cfg.Level, "debug")
	}

	if cfg.Format != textFormat {
		t.Errorf("Config.Format = %q, want %q", cfg.Format, textFormat)
	}

	if cfg.Output == nil {
		t.Error("Config.Output should not be nil")
	}
}

func TestLoggerOutput(t *testing.T) {
	tests := []struct {
		name   string
		format string
		level  string
	}{
		{
			name:   "json format produces valid output",
			format: "json",
			level:  "info",
		},
		{
			name:   "text format produces valid output",
			format: "text",
			level:  "debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			if tt.format == "json" && !strings.Contains(output, `"msg":"test message"`) {
				t.Error("json format should contain structured message")
			}
			if tt.format == "text" && !strings.Contains(output, "test message") {
				t.Error("text format should contain readable message")
			}
		})
	}
}