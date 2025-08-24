package env

import (
	"os"
	"strings"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		envValue string
		want     string
	}{
		{
			name:     "returns environment variable when set",
			key:      "TEST_VAR",
			fallback: "default",
			envValue: "test_value",
			want:     "test_value",
		},
		{
			name:     "returns fallback when environment variable not set",
			key:      "UNSET_VAR",
			fallback: "default_value",
			envValue: "",
			want:     "default_value",
		},
		{
			name:     "returns fallback when environment variable is empty",
			key:      "EMPTY_VAR",
			fallback: "fallback",
			envValue: "",
			want:     "fallback",
		},
		{
			name:     "handles lowercase key conversion",
			key:      "lowercase_key",
			fallback: "default",
			envValue: "uppercase_value",
			want:     "uppercase_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			envKey := EnvPrefix + strings.ToUpper(tt.key)
			if tt.envValue != "" {
				_ = os.Setenv(envKey, tt.envValue)
				defer func() { _ = os.Unsetenv(envKey) }()
			} else {
				_ = os.Unsetenv(envKey) // Ensure it's not set
			}

			// Test
			got := GetEnv(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("GetEnv(%q, %q) = %q, want %q", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}


func TestEnvPrefix(t *testing.T) {
	t.Run("prefix constant is correct", func(t *testing.T) {
		const expectedPrefix = "SCRY_"
		if EnvPrefix != expectedPrefix {
			t.Errorf("EnvPrefix = %q, want %q", EnvPrefix, expectedPrefix)
		}
	})
}

func TestGetEnvKeyTransformation(t *testing.T) {
	t.Run("keys are properly converted to uppercase", func(t *testing.T) {
		testKey := "log_format"
		expectedEnvKey := "SCRY_LOG_FORMAT"
		
		_ = os.Setenv(expectedEnvKey, "test_value")
		defer func() { _ = os.Unsetenv(expectedEnvKey) }()
		
		got := GetEnv(testKey, "default")
		if got != "test_value" {
			t.Errorf("GetEnv(%q, %q) = %q, want %q", testKey, "default", got, "test_value")
		}
	})
}