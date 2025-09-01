package embeddings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelDimensions(t *testing.T) {
	t.Run("chat model dimensions", func(t *testing.T) {
		assert.Equal(t, 2880, ModelDimension(Chat))
	})

	t.Run("embedding model dimensions", func(t *testing.T) {
		assert.Equal(t, 768, ModelDimension(Embedding))
	})

	t.Run("unknown model defaults to chat dimensions", func(t *testing.T) {
		assert.Equal(t, 2880, ModelDimension(Model("unknown-model")))
	})
}
