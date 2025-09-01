package embeddings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client successfully", func(t *testing.T) {
		cfg := Config{
			Host:  "http://localhost:11434",
			Model: Chat,
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, cfg.Model, client.config.Model)
	})
}

func TestGenerateEmbedding(t *testing.T) {
	t.Run("generates embedding with correct dimensions", func(t *testing.T) {
		t.Skip("Integration test - requires Ollama server")

		cfg := Config{
			Host:  "http://localhost:11434",
			Model: Embedding,
		}

		client, err := NewClient(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		embedding, err := client.GenerateEmbedding(ctx, "test text")
		require.NoError(t, err)
		assert.NotEmpty(t, embedding)
		assert.Len(t, embedding, ModelDimension(Embedding))
	})
}
