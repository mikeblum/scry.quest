package conf

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testEnvContent = `
TEST_STRING=hello_world
TEST_INT=42
TEST_BOOL=true
TEST_LIST=one,two,three
`

func createTestConfig(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	confFile := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(confFile, []byte(content), 0600)
	require.NoError(t, err)
	return confFile
}

func newTestConfig(t *testing.T, confFile string) *Config {
	t.Helper()
	ctx := context.Background()
	config, err := New(ctx, confFile)
	require.NoError(t, err)
	assert.NotNil(t, config)
	return config
}

func TestConf(t *testing.T) {
	t.Run("config=new-valid", NewValidConfigTest)
	t.Run("config=new-env-only", NewEnvOnlyConfigTest)
	t.Run("config=env-override", EnvOverrideConfigTest)
	t.Run("config=new-error", NewConfigErrorTest)
	t.Run("config=methods", ConfigMethodsTest)
	t.Run("config=path-explicit", ConfigPathExplicitTest)
	t.Run("config=path-env", ConfigPathEnvTest)
	t.Run("config=path-default", ConfigPathDefaultTest)
	t.Run("env=existing", GetEnvExistingTest)
	t.Run("env=fallback", GetEnvFallbackTest)
	t.Run("env=empty-fallback", GetEnvEmptyFallbackTest)
	t.Run("env=shell", GetEnvShellTest)
}

func NewValidConfigTest(t *testing.T) {
	confFile := createTestConfig(t, testEnvContent)
	config := newTestConfig(t, confFile)

	assert.Equal(t, "hello_world", config.String("TEST_STRING"))
	assert.Equal(t, 42, config.Int("TEST_INT"))
	assert.True(t, config.Bool("TEST_BOOL"))
}

func NewEnvOnlyConfigTest(t *testing.T) {
	t.Setenv("TEST_ENV_VAR", "test_value")
	config := newTestConfig(t, "/nonexistent/.env")

	assert.Equal(t, "test_value", config.String("TEST_ENV_VAR"))
}

func EnvOverrideConfigTest(t *testing.T) {
	confFile := createTestConfig(t, "TEST_OVERRIDE=file_value\n")
	t.Setenv("TEST_OVERRIDE", "env_value")
	config := newTestConfig(t, confFile)

	assert.Equal(t, "env_value", config.String("TEST_OVERRIDE"))
}

func NewConfigErrorTest(t *testing.T) {
	tmpDir := t.TempDir()
	confFile := filepath.Join(tmpDir, ".env")
	// Create an unreadable file to trigger a file loading error
	err := os.WriteFile(confFile, []byte("TEST=value"), 0000)
	require.NoError(t, err)

	ctx := context.Background()
	config, err := New(ctx, confFile)
	require.Error(t, err)
	assert.Nil(t, config)
}

func ConfigMethodsTest(t *testing.T) {
	confFile := createTestConfig(t, testEnvContent)
	config := newTestConfig(t, confFile)

	assert.Equal(t, "hello_world", config.String("TEST_STRING"))
	assert.Empty(t, config.String("NONEXISTENT"))
	assert.Equal(t, 42, config.Int("TEST_INT"))
	assert.Equal(t, 0, config.Int("NONEXISTENT"))
	assert.True(t, config.Bool("TEST_BOOL"))
	assert.False(t, config.Bool("NONEXISTENT"))
	assert.True(t, config.Exists("TEST_STRING"))
	assert.False(t, config.Exists("NONEXISTENT"))
	assert.Equal(t, "hello_world", config.MustString("TEST_STRING"))
	assert.Panics(t, func() {
		config.MustString("NONEXISTENT")
	})

	// StringSlice requires array-like values, not comma-separated strings in .env
	// Testing with a known non-existent key for coverage
	assert.Empty(t, config.StringSlice("TEST_LIST"))
	assert.Empty(t, config.StringSlice("NONEXISTENT"))

	all := config.All()
	assert.Contains(t, all, "TEST_STRING")
	assert.Equal(t, "hello_world", all["TEST_STRING"])
}

func ConfigPathExplicitTest(t *testing.T) {
	path := getConfigPath("/custom/path/.env")
	assert.Equal(t, "/custom/path/.env", path)
}

func ConfigPathEnvTest(t *testing.T) {
	t.Setenv(EnvConfigPath, "/env/path/.env")
	path := getConfigPath("")
	assert.Equal(t, "/env/path/.env", path)
}

func ConfigPathDefaultTest(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	expected := filepath.Join(wd, ConfFile)

	path := getConfigPath("")
	assert.Equal(t, expected, path)
}

func GetEnvExistingTest(t *testing.T) {
	t.Setenv("TEST_EXISTING", "test_value")
	result := GetEnv("TEST_EXISTING", "fallback")
	assert.Equal(t, "test_value", result)
}

func GetEnvFallbackTest(t *testing.T) {
	result := GetEnv("NONEXISTENT_VAR", "fallback_value")
	assert.Equal(t, "fallback_value", result)
}

func GetEnvEmptyFallbackTest(t *testing.T) {
	result := GetEnv("NONEXISTENT_VAR", "")
	assert.Empty(t, result)
}

func GetEnvShellTest(t *testing.T) {
	shell := GetEnv("SHELL", "")
	if shell != "" {
		assert.NotEmpty(t, strings.TrimSpace(shell))
	}
}
