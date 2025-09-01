package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestEmbeddingsApp(t *testing.T) {
	t.Run("creates embeddings app successfully", func(t *testing.T) {
		app := NewEmbeddingsCLI()
		require.NotNil(t, app)
		assert.Equal(t, "embeddings", app.Name)
		assert.Equal(t, "SRD embeddings pipeline", app.Usage)
	})

	t.Run("has all expected commands", func(t *testing.T) {
		app := NewEmbeddingsCLI()
		require.NotNil(t, app)

		expectedCommands := []string{"migrate", "generate", "search", "stats", "clear"}
		require.Len(t, app.Commands, len(expectedCommands))

		commandNames := make([]string, len(app.Commands))
		for i, cmd := range app.Commands {
			commandNames[i] = cmd.Name
		}

		for _, expectedCmd := range expectedCommands {
			assert.Contains(t, commandNames, expectedCmd)
		}
	})

	t.Run("has all expected global flags", func(t *testing.T) {
		app := NewEmbeddingsCLI()
		require.NotNil(t, app)

		expectedFlags := []string{"config", "model", "ollama-url"}
		require.Len(t, app.Flags, len(expectedFlags))

		flagNames := make([]string, len(app.Flags))
		for i, flag := range app.Flags {
			flagNames[i] = flag.Names()[0]
		}

		for _, expectedFlag := range expectedFlags {
			assert.Contains(t, flagNames, expectedFlag)
		}
	})

	t.Run("help command executes without error", func(t *testing.T) {
		app := NewEmbeddingsCLI()
		require.NotNil(t, app)

		args := []string{"embeddings", "--help"}
		err := app.RunContext(context.Background(), args)
		assert.NoError(t, err)
	})

	t.Run("has required structure", func(t *testing.T) {
		app := NewEmbeddingsCLI()
		require.NotNil(t, app)

		assert.NotEmpty(t, app.Name)
		assert.NotEmpty(t, app.Usage)
		assert.NotEmpty(t, app.Commands)
		assert.NotEmpty(t, app.Flags)
	})
}

func TestCommands(t *testing.T) {
	t.Run("generate command has expected flags", func(t *testing.T) {
		app := NewEmbeddingsCLI()
		require.NotNil(t, app)

		var generateCmd *cli.Command
		for _, cmd := range app.Commands {
			if cmd.Name == "generate" {
				generateCmd = cmd
				break
			}
		}

		require.NotNil(t, generateCmd)
		assert.Equal(t, "Generate embeddings for SRD content", generateCmd.Usage)

		var hasTypeFlag bool
		for _, flag := range generateCmd.Flags {
			if flag.Names()[0] == "type" {
				hasTypeFlag = true
				break
			}
		}
		assert.True(t, hasTypeFlag)
	})

	t.Run("search command has expected flags", func(t *testing.T) {
		app := NewEmbeddingsCLI()
		require.NotNil(t, app)

		var searchCmd *cli.Command
		for _, cmd := range app.Commands {
			if cmd.Name == "search" {
				searchCmd = cmd
				break
			}
		}

		require.NotNil(t, searchCmd)
		assert.Equal(t, "Search content using embeddings", searchCmd.Usage)

		expectedFlags := []string{"query", "type", "limit"}
		flagNames := make([]string, len(searchCmd.Flags))
		for i, flag := range searchCmd.Flags {
			flagNames[i] = flag.Names()[0]
		}

		for _, expectedFlag := range expectedFlags {
			assert.Contains(t, flagNames, expectedFlag)
		}
	})
}
