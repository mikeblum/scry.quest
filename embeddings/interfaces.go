package embeddings

import (
	"context"

	"github.com/google/uuid"
)

// ContentItem represents content for processing.
type ContentItem struct {
	ID       uuid.UUID
	Content  []byte
	Metadata map[string]interface{}
	Type     string
}

// EmbeddingResult represents processed content with embeddings.
type EmbeddingResult struct {
	ContentID   uuid.UUID
	Embedding   []float32
	ContentType string
	Metadata    map[string]interface{}
}

// DataSource provides ContentItems.
type DataSource interface {
	Read(ctx context.Context) (<-chan *ContentItem, error)
	Close() error
}

// ContentProcessor transforms ContentItems to EmbeddingResults.
type ContentProcessor interface {
	Process(ctx context.Context, item *ContentItem) (*EmbeddingResult, error)
}

// EmbeddingStore persists EmbeddingResults.
type EmbeddingStore interface {
	Store(ctx context.Context, result *EmbeddingResult) error
	StoreAll(ctx context.Context, results []*EmbeddingResult) error
}

// EmbeddingGenerator creates embeddings from text.
type EmbeddingGenerator interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}
