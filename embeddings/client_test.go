package embeddings //nolint:revive // package comment not needed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	cfg := Config{
		Host:  "http://localhost:11434",
		Model: "gpt-oss:20b",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, cfg.Model, client.config.Model)
}

func TestGetModelDimensions(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected int
	}{
		{"gpt-oss:20b", "gpt-oss:20b", 1536},
		{"nomic-embed-text", "nomic-embed-text", 768},
		{"unknown", "unknown-model", 1536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Host:  "http://localhost:11434",
				Model: tt.model,
			}
			client, err := NewClient(cfg)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, client.GetModelDimensions())
		})
	}
}

func TestGenerateEmbedding(t *testing.T) {
	t.Skip("Integration test - requires Ollama server")

	cfg := Config{
		Host:  "http://localhost:11434",
		Model: "nomic-embed-text",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	embedding, err := client.GenerateEmbedding(ctx, "test text")
	require.NoError(t, err)
	assert.NotEmpty(t, embedding)
	assert.Len(t, embedding, 768)
}
