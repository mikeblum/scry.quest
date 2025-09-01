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

	embedding, err := p.generator.GenerateEmbedding(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	result := &EmbeddingResult{
		ContentID:   item.ID,
		Embedding:   embedding,
		ContentType: item.Type,
		Metadata:    make(map[string]interface{}),
	}

	// Copy metadata from item
	for k, v := range item.Metadata {
		result.Metadata[k] = v
	}

	// Store the original content
	result.Metadata["original_content"] = item.Content

	// Add processor metadata
	result.Metadata["model"] = string(p.model)
	result.Metadata["embedding_dimensions"] = len(embedding)

	return result, nil
}
