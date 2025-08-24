// Package conf provides configuration management using koanf with support for
// .env files and environment variables.
package conf

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

const (
	// EnvConfigPath is the environment variable name used to specify a custom config file path
	EnvConfigPath   = "CONF_PATH"
	// EnvVarNamespace is the namespace prefix for environment variables
	EnvVarNamespace = ""
	// EnvDelimiter is the delimiter used for environment variable keys
	EnvDelimiter    = "."
	// PropDelimiter is the delimiter used for property keys
	PropDelimiter   = "."
	// ConfFile is the default configuration file name
	ConfFile        = ".env"
)

// Config wraps koanf configuration management with additional convenience methods
type Config struct {
	koanf *koanf.Koanf
}

// New creates a new Config instance by loading configuration from .env file and environment variables
func New(ctx context.Context, configPath string) (*Config, error) {
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

// Int returns the integer value for the given key
func (c *Config) Int(key string) int {
	return c.koanf.Int(key)
}

// Bool returns the boolean value for the given key
func (c *Config) Bool(key string) bool {
	return c.koanf.Bool(key)
}

// StringSlice returns a string slice for the given key
func (c *Config) StringSlice(key string) []string {
	return c.koanf.Strings(key)
}

// Exists returns true if the key exists in the configuration
func (c *Config) Exists(key string) bool {
	return c.koanf.Exists(key)
}

// All returns all configuration key-value pairs
func (c *Config) All() map[string]interface{} {
	return c.koanf.All()
}

// MustString returns the string value for the given key or panics if not found
func (c *Config) MustString(key string) string {
	if !c.koanf.Exists(key) {
		panic(fmt.Sprintf("required configuration key %q not found", key))
	}
	return c.koanf.String(key)
}

func getConfigPath(configPath string) string {
	if configPath != "" {
		return configPath
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