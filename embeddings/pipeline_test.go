package embeddings

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDataSource struct {
	items  []*ContentItem
	closed bool
}

func (m *mockDataSource) Read(ctx context.Context) (<-chan *ContentItem, error) {
	items := make(chan *ContentItem)
	go func() {
		defer close(items)
		for _, item := range m.items {
			select {
			case <-ctx.Done():
				return
			case items <- item:
			}
		}
	}()
	return items, nil
}

func (m *mockDataSource) Close() error {
	m.closed = true
	return nil
}

type mockContentProcessor struct {
	processedItems []*ContentItem
}

func (m *mockContentProcessor) Process(_ context.Context, item *ContentItem) (*EmbeddingResult, error) {
	m.processedItems = append(m.processedItems, item)
	return &EmbeddingResult{
		ContentID:   item.ID,
		Embedding:   []float32{0.1, 0.2, 0.3},
		ContentType: item.Type,
		Metadata:    map[string]interface{}{"processed": true},
	}, nil
}

type mockEmbeddingStore struct {
	storedResults []*EmbeddingResult
}

func (m *mockEmbeddingStore) Store(_ context.Context, result *EmbeddingResult) error {
	m.storedResults = append(m.storedResults, result)
	return nil
}

func (m *mockEmbeddingStore) StoreAll(_ context.Context, results []*EmbeddingResult) error {
	m.storedResults = append(m.storedResults, results...)
	return nil
}

func TestGenericPipeline(t *testing.T) {
	t.Run("processes source successfully", func(t *testing.T) {
		generator := &mockEmbeddingGenerator{
			embeddings: [][]float32{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}},
		}
		processor := &mockContentProcessor{}
		store := &mockEmbeddingStore{}

		pipeline := NewGenericPipeline(generator, processor, store)

		source := &mockDataSource{
			items: []*ContentItem{
				{ID: uuid.New(), Content: []byte("content 1"), Type: "test"},
				{ID: uuid.New(), Content: []byte("content 2"), Type: "test"},
			},
		}

		ctx := context.Background()
		err := pipeline.ProcessSource(ctx, source)
		require.NoError(t, err)

		assert.True(t, source.closed)
		assert.Len(t, processor.processedItems, 2)
		assert.Len(t, store.storedResults, 2)

		assert.NotEmpty(t, processor.processedItems[0].ID)
		assert.NotEmpty(t, processor.processedItems[1].ID)
	})

	t.Run("processes single item successfully", func(t *testing.T) {
		generator := &mockEmbeddingGenerator{
			embeddings: [][]float32{{0.1, 0.2, 0.3}},
		}
		processor := &mockContentProcessor{}
		store := &mockEmbeddingStore{}

		pipeline := NewGenericPipeline(generator, processor, store)

		item := &ContentItem{
			ID:      uuid.New(),
			Content: []byte("content 1"),
			Type:    "test",
		}

		ctx := context.Background()
		result, err := pipeline.ProcessItem(ctx, item)
		require.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, item.ID, result.ContentID)
		assert.Len(t, processor.processedItems, 1)
		assert.Len(t, store.storedResults, 1)
	})
}
