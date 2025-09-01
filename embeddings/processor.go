package embeddings

import (
	"context"
	"fmt"
)

// DefaultContentProcessor processes content using EmbeddingGenerator.
type DefaultContentProcessor struct {
	generator EmbeddingGenerator
	model     Model
}

// NewDefaultContentProcessor creates a content processor.
func NewDefaultContentProcessor(generator EmbeddingGenerator, model Model) *DefaultContentProcessor {
	return &DefaultContentProcessor{
		generator: generator,
		model:     model,
	}
}

// Process generates embeddings for content.
func (p *DefaultContentProcessor) Process(ctx context.Context, item *ContentItem) (*EmbeddingResult, error) {
	if item == nil {
		return nil, fmt.Errorf("content item cannot be nil")
	}

	text := string(item.Content)
	if text == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	// Respect model context length
	maxLength := ModelContextLength(p.model)
	if len(text) > maxLength {
		text = text[:maxLength]
	}

	embedding, err := p.generator.GenerateEmbedding(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Simple result - copy original metadata and add processor metadata
	result := &EmbeddingResult{
		ContentID:   item.ID,
		Embedding:   embedding,
		ContentType: item.Type,
		Metadata:    make(map[string]any),
	}

	// Copy all original metadata
	for k, v := range item.Metadata {
		result.Metadata[k] = v
	}

	// Add processor metadata
	result.Metadata["original_content"] = item.Content
	result.Metadata["model"] = string(p.model)
	result.Metadata["embedding_dimensions"] = len(embedding)

	return result, nil
}
