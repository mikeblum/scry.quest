package main

import (
	"context"
	"testing"

	"github.com/mikeblum/scry.quest/embeddings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestAppCreation(t *testing.T) {
	engine := embeddings.NewEngine()
	require.NotNil(t, engine)
	assert.Equal(t, "embeddings", engine.Name)
	assert.Equal(t, "D&D 5e SRD embeddings pipeline", engine.Usage)
}

func TestAppCommands(t *testing.T) {
	engine := embeddings.NewEngine()
	require.NotNil(t, engine)

	expectedCommands := []string{"generate", "search", "stats", "clear"}
	require.Len(t, engine.Commands, len(expectedCommands))

	commandNames := make([]string, len(engine.Commands))
	for i, cmd := range engine.Commands {
		commandNames[i] = cmd.Name
	}

	for _, expectedCmd := range expectedCommands {
		assert.Contains(t, commandNames, expectedCmd)
	}
}

func TestAppFlags(t *testing.T) {
	engine := embeddings.NewEngine()
	require.NotNil(t, engine)

	expectedFlags := []string{"config", "model", "ollama-url"}
	require.Len(t, engine.Flags, len(expectedFlags))

	flagNames := make([]string, len(engine.Flags))
	for i, flag := range engine.Flags {
		flagNames[i] = flag.Names()[0]
	}

	for _, expectedFlag := range expectedFlags {
		assert.Contains(t, flagNames, expectedFlag)
	}
}

func TestAppHelpCommand(t *testing.T) {
	engine := embeddings.NewEngine()
	require.NotNil(t, engine)

	// Test help command doesn't error
	args := []string{"embeddings", "--help"}
	err := engine.RunContext(context.Background(), args)
	// Help command should not return an error in urfave/cli
	assert.NoError(t, err)
}

func TestAppBasicValidation(t *testing.T) {
	engine := embeddings.NewEngine()
	require.NotNil(t, engine)

	// Verify engine has the expected structure
	assert.NotEmpty(t, engine.Name)
	assert.NotEmpty(t, engine.Usage)
	assert.NotEmpty(t, engine.Commands)
	assert.NotEmpty(t, engine.Flags)
}

func TestGenerateCommandFlags(t *testing.T) {
	engine := embeddings.NewEngine()
	require.NotNil(t, engine)

	var generateCmd *cli.Command
	for _, cmd := range engine.Commands {
		if cmd.Name == "generate" {
			generateCmd = cmd
			break
		}
	}

	require.NotNil(t, generateCmd)
	assert.Equal(t, "Generate embeddings for SRD content", generateCmd.Usage)

	// Check for type flag
	var hasTypeFlag bool
	for _, flag := range generateCmd.Flags {
		if flag.Names()[0] == "type" {
			hasTypeFlag = true
			break
		}
	}
	assert.True(t, hasTypeFlag)
}

func TestSearchCommandFlags(t *testing.T) {
	engine := embeddings.NewEngine()
	require.NotNil(t, engine)

	var searchCmd *cli.Command
	for _, cmd := range engine.Commands {
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
}
