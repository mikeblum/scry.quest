package embeddings

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEmbeddingGenerator struct {
	embeddings [][]float32
	callCount  int
}

func (m *mockEmbeddingGenerator) GenerateEmbedding(_ context.Context, _ string) ([]float32, error) {
	if m.callCount >= len(m.embeddings) {
		return []float32{0.1, 0.2, 0.3}, nil // default embedding
	}
	embedding := m.embeddings[m.callCount]
	m.callCount++
	return embedding, nil
}

func TestDefaultContentProcessor(t *testing.T) {
	t.Run("processes content item successfully", func(t *testing.T) {
		generator := &mockEmbeddingGenerator{
			embeddings: [][]float32{{0.1, 0.2, 0.3}},
		}
		processor := NewDefaultContentProcessor(generator, Chat)

		item := &ContentItem{
			ID:      uuid.New(),
			Content: []byte("test content"),
			Type:    "test",
			Metadata: map[string]interface{}{
				"source": "test",
			},
		}

		ctx := context.Background()
		result, err := processor.Process(ctx, item)
		require.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, item.ID, result.ContentID)
		assert.Equal(t, "test", result.ContentType)
		assert.Equal(t, []float32{0.1, 0.2, 0.3}, result.Embedding)
		assert.Equal(t, string(Chat), result.Metadata["model"])
		assert.Equal(t, 3, result.Metadata["embedding_dimensions"])
		assert.Equal(t, "test", result.Metadata["source"])
		assert.Equal(t, []byte("test content"), result.Metadata["original_content"])
	})

	t.Run("handles nil content item", func(t *testing.T) {
		generator := &mockEmbeddingGenerator{}
		processor := NewDefaultContentProcessor(generator, Chat)

		ctx := context.Background()
		result, err := processor.Process(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "content item cannot be nil")
	})

	t.Run("handles empty content", func(t *testing.T) {
		generator := &mockEmbeddingGenerator{}
		processor := NewDefaultContentProcessor(generator, Chat)

		item := &ContentItem{
			ID:      uuid.New(),
			Content: []byte(""),
			Type:    "test",
		}

		ctx := context.Background()
		result, err := processor.Process(ctx, item)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "content cannot be empty")
	})
}
