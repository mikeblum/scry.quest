package embeddings

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Pipeline processes content through pluggable components.
type Pipeline struct {
	generator EmbeddingGenerator
	processor ContentProcessor
	store     EmbeddingStore
}

// NewPipeline creates a processing pipeline.
func NewPipeline(generator EmbeddingGenerator, processor ContentProcessor, store EmbeddingStore) *Pipeline {
	return &Pipeline{
		generator: generator,
		processor: processor,
		store:     store,
	}
}

// ProcessSource processes DataSource items through pipeline.
func (p *Pipeline) ProcessSource(ctx context.Context, source DataSource) error {
	defer func() {
		if err := source.Close(); err != nil {
			slog.ErrorContext(ctx, "Failed to close data source", "error", err)
		}
	}()

	items, err := source.Read(ctx)
	if err != nil {
		return fmt.Errorf("failed to read from data source: %w", err)
	}

	results := make([]*EmbeddingResult, 0)
	for item := range items {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		start := time.Now().UTC()
		result, err := p.processor.Process(ctx, item)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to process content item",
				"id", item.ID, "type", item.Type, "error", err)
			continue
		}

		results = append(results, result)

		logFields := []interface{}{"id", item.ID, "type", item.Type, "duration", time.Since(start)}
		if name := extractName(item); name != nil {
			logFields = append(logFields, "name", *name)
		}
		if filePath := extractFilePath(item); filePath != nil {
			logFields = append(logFields, "file", *filePath)
		}
		logFields = append(logFields, "vector_size", len(result.Embedding))
		slog.InfoContext(ctx, "Processed SRD item", logFields...)
	}

	if len(results) > 0 {
		if err := p.store.StoreAll(ctx, results); err != nil {
			return fmt.Errorf("failed to store results: %w", err)
		}
	}

	return nil
}

// ProcessItem processes a single ContentItem.
func (p *Pipeline) ProcessItem(ctx context.Context, item *ContentItem) (*EmbeddingResult, error) {
	result, err := p.processor.Process(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to process item %s: %w", item.ID, err)
	}

	if err := p.store.Store(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to store result for item %s: %w", item.ID, err)
	}

	return result, nil
}

func extractName(item *ContentItem) *string {
	if nameVal, ok := item.Metadata["name"]; ok {
		if nameStr, ok := nameVal.(string); ok {
			return &nameStr
		}
	}
	return nil
}

func extractFilePath(item *ContentItem) *string {
	if pathVal, ok := item.Metadata["file_path"]; ok {
		if pathStr, ok := pathVal.(string); ok {
			return &pathStr
		}
	}
	return nil
}
