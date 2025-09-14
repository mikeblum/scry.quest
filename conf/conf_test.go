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
	config, err := New(ctx, &confFile)
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
	t.Run("prefixed-env=config-exists", GetPrefixedEnvFromConfigTest)
	t.Run("prefixed-env=env-exists", GetPrefixedEnvFromEnvTest)
	t.Run("prefixed-env=fallback", GetPrefixedEnvFallbackTest)
	t.Run("prefixed-env=config-override-env", GetPrefixedEnvConfigOverrideTest)
	t.Run("cli-fallback=flag-priority", GetWithCLIFallbackFlagPriorityTest)
	t.Run("cli-fallback=env-fallback", GetWithCLIFallbackEnvFallbackTest)
	t.Run("cli-fallback=default-fallback", GetWithCLIFallbackDefaultTest)
	t.Run("cli-fallback=nil-context", GetWithCLIFallbackNilContextTest)
}

func NewValidConfigTest(t *testing.T) {
	confFile := createTestConfig(t, testEnvContent)
	config := newTestConfig(t, confFile)

	assert.Equal(t, "hello_world", config.String("TEST_STRING"))
	assert.Equal(t, 42, config.koanf.Int("TEST_INT"))
	assert.True(t, config.koanf.Bool("TEST_BOOL"))
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
	config, err := New(ctx, &confFile)
	require.Error(t, err)
	assert.Nil(t, config)
}

func ConfigMethodsTest(t *testing.T) {
	confFile := createTestConfig(t, testEnvContent)
	config := newTestConfig(t, confFile)

	assert.Equal(t, "hello_world", config.String("TEST_STRING"))
	assert.Empty(t, config.String("NONEXISTENT"))
	assert.Equal(t, 42, config.koanf.Int("TEST_INT"))
	assert.Equal(t, 0, config.koanf.Int("NONEXISTENT"))
	assert.True(t, config.koanf.Bool("TEST_BOOL"))
	assert.False(t, config.koanf.Bool("NONEXISTENT"))
	assert.True(t, config.koanf.Exists("TEST_STRING"))
	assert.False(t, config.koanf.Exists("NONEXISTENT"))
	value, err := config.MustString("TEST_STRING")
	require.NoError(t, err)
	assert.Equal(t, "hello_world", value)
	_, err = config.MustString("NONEXISTENT")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required configuration key \"NONEXISTENT\" not found")

	// StringSlice requires array-like values, not comma-separated strings in .env
	// Testing with a known non-existent key for coverage
	assert.Empty(t, config.koanf.Strings("TEST_LIST"))
	assert.Empty(t, config.koanf.Strings("NONEXISTENT"))

	all := config.koanf.All()
	assert.Contains(t, all, "TEST_STRING")
	assert.Equal(t, "hello_world", all["TEST_STRING"])
}

func ConfigPathExplicitTest(t *testing.T) {
	customPath := "/custom/path/.env"
	path := getConfigPath(&customPath)
	assert.Equal(t, "/custom/path/.env", path)
}

func ConfigPathEnvTest(t *testing.T) {
	t.Setenv(EnvConfigPath, "/env/path/.env")
	emptyPath := ""
	path := getConfigPath(&emptyPath)
	assert.Equal(t, "/env/path/.env", path)
}

func ConfigPathDefaultTest(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	expected := filepath.Join(wd, ConfFile)

	path := getConfigPath(nil)
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

func GetPrefixedEnvFromConfigTest(t *testing.T) {
	confFile := createTestConfig(t, "SCRY_TEST_KEY=config_value\n")
	config := newTestConfig(t, confFile)

	result := config.GetPrefixedEnv("TEST_KEY", "fallback")
	assert.Equal(t, "config_value", result)
}

func GetPrefixedEnvFromEnvTest(t *testing.T) {
	t.Setenv("SCRY_TEST_ENV", "env_value")
	config := newTestConfig(t, "/nonexistent/.env")

	result := config.GetPrefixedEnv("TEST_ENV", "fallback")
	assert.Equal(t, "env_value", result)
}

func GetPrefixedEnvFallbackTest(t *testing.T) {
	config := newTestConfig(t, "/nonexistent/.env")

	result := config.GetPrefixedEnv("NONEXISTENT_KEY", "fallback_value")
	assert.Equal(t, "fallback_value", result)
}

func GetPrefixedEnvConfigOverrideTest(t *testing.T) {
	confFile := createTestConfig(t, "SCRY_OVERRIDE_TEST=config_value\n")
	t.Setenv("SCRY_OVERRIDE_TEST", "env_value")
	config := newTestConfig(t, confFile)

	result := config.GetPrefixedEnv("OVERRIDE_TEST", "fallback")
	assert.Equal(t, "env_value", result)
}

// Mock CLI context for testing
type mockCLIContext struct {
	values map[string]string
}

func (m *mockCLIContext) String(key string) string {
	if m.values == nil {
		return ""
	}
	return m.values[key]
}

func GetWithCLIFallbackFlagPriorityTest(t *testing.T) {
	config := newTestConfig(t, "/nonexistent/.env")

	cliContext := &mockCLIContext{
		values: map[string]string{
			"test-flag": "cli_value",
		},
	}

	result := config.FromCLI(cliContext, "test-flag", "TEST_ENV", "default")
	assert.Equal(t, "cli_value", result)
}

func GetWithCLIFallbackEnvFallbackTest(t *testing.T) {
	t.Setenv("SCRY_TEST_ENV", "env_value")
	config := newTestConfig(t, "/nonexistent/.env")

	cliContext := &mockCLIContext{
		values: map[string]string{},
	}

	result := config.FromCLI(cliContext, "test-flag", "TEST_ENV", "default")
	assert.Equal(t, "env_value", result)
}

func GetWithCLIFallbackDefaultTest(t *testing.T) {
	config := newTestConfig(t, "/nonexistent/.env")

	cliContext := &mockCLIContext{
		values: map[string]string{},
	}

	result := config.FromCLI(cliContext, "test-flag", "NONEXISTENT_ENV", "default_value")
	assert.Equal(t, "default_value", result)
}

func GetWithCLIFallbackNilContextTest(t *testing.T) {
	t.Setenv("SCRY_TEST_ENV", "env_value")
	config := newTestConfig(t, "/nonexistent/.env")

	result := config.FromCLI(nil, "test-flag", "TEST_ENV", "default")
	assert.Equal(t, "env_value", result)
}
