// Package env provides utilities for working with environment variables
package env

import (
	"os"
	"strings"
)

// EnvPrefix is the prefix used for all environment variables
const EnvPrefix = "SCRY_"

// GetEnv gets environment variable with SCRY_ prefix and fallback
func GetEnv(key, fallback string) string {
	envKey := EnvPrefix + strings.ToUpper(key)
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return fallback
}
