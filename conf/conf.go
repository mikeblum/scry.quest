// Package conf provides configuration management using koanf with support for
// .env files and environment variables.
package conf

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	// EnvConfigPath is the environment variable name used to specify a custom config file path
	EnvConfigPath = "CONF_PATH"
	// EnvPrefix to scope env vars to scry_quest. See [env.template](../.env.template) for examples
	EnvPrefix = "SCRY_"
	// EnvVarNamespace is the namespace prefix for environment variables
	EnvVarNamespace = ""
	// EnvDelimiter is the delimiter used for environment variable keys
	EnvDelimiter = "."
	// PropDelimiter is the delimiter used for property keys
	PropDelimiter = "."
	// ConfFile is the default configuration file name
	ConfFile = ".env"
)

// Config wraps koanf configuration management with additional convenience methods
type Config struct {
	koanf *koanf.Koanf
}

// New creates a new Config instance by loading configuration from .env file and environment variables
func New(ctx context.Context, configPath *string) (*Config, error) {
	k := koanf.New(".")

	confFile := getConfigPath(configPath)

	if _, err := os.Stat(confFile); err == nil {
		slog.InfoContext(ctx, "loading configuration", "file", confFile)
		if err := k.Load(file.Provider(confFile), dotenv.Parser()); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", confFile, err)
		}
	} else {
		slog.WarnContext(ctx, "configuration file not found, using environment variables only", "file", confFile)
	}

	if err := k.Load(env.Provider(EnvVarNamespace, EnvDelimiter, func(s string) string { return s }), nil); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	return &Config{koanf: k}, nil
}

// String returns the string value for the given key
func (c *Config) String(key string) string {
	return c.koanf.String(key)
}

// MustString returns the string value for the given key or panics if not found
func (c *Config) MustString(key string) string {
	if !c.koanf.Exists(key) {
		panic(fmt.Sprintf("required configuration key %q not found", key))
	}
	return c.koanf.String(key)
}

func getConfigPath(configPath *string) string {
	if configPath != nil && *configPath != "" {
		return *configPath
	}

	if path := os.Getenv(EnvConfigPath); path != "" {
		return path
	}

	if wd, err := os.Getwd(); err == nil {
		return filepath.Join(wd, ConfFile)
	}

	return ConfFile
}

// GetEnv returns the environment variable value or the fallback if not set
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// GetPrefixedEnv returns the configuration value with SCRY_ prefix from config or env, or fallback if not set
func (c *Config) GetPrefixedEnv(key, fallback string) string {
	prefixedKey := EnvPrefix + key

	// First check if it exists in koanf (from .env file or environment)
	if c.koanf.Exists(prefixedKey) {
		return c.koanf.String(prefixedKey)
	}

	// Fallback to direct environment lookup
	if value, exists := os.LookupEnv(prefixedKey); exists {
		return value
	}

	return fallback
}

// CLIContext represents a CLI context that can return string flag values
type CLIContext interface {
	String(name string) string
}

// FromCLI resolves a configuration value with the following priority:
// 1. CLI context flag value (if provided)
// 2. Configuration/environment value via GetPrefixedEnv
// 3. Default fallback value
func (c *Config) FromCLI(cliContext CLIContext, flagName, envKey, fallback string) string {
	// Try CLI flag first if context is provided
	if cliContext != nil {
		if flagValue := cliContext.String(flagName); flagValue != "" {
			return flagValue
		}
	}

	// Fall back to prefixed environment variable or config
	return c.GetPrefixedEnv(envKey, fallback)
}
