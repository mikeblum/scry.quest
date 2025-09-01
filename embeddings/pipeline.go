package embeddings

import (
	"context"
	"fmt"
	"log/slog"
)

// GenericPipeline processes content through pluggable components.
type GenericPipeline struct {
	generator EmbeddingGenerator
	processor ContentProcessor
	store     EmbeddingStore
}

// NewGenericPipeline creates a processing pipeline.
func NewGenericPipeline(generator EmbeddingGenerator, processor ContentProcessor, store EmbeddingStore) *GenericPipeline {
	return &GenericPipeline{
		generator: generator,
		processor: processor,
		store:     store,
	}
}

// ProcessSource processes DataSource items through pipeline.
func (p *GenericPipeline) ProcessSource(ctx context.Context, source DataSource) error {
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result, err := p.processor.Process(ctx, item)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to process content item",
				"id", item.ID, "type", item.Type, "error", err)
			continue
		}

		results = append(results, result)

		logFields := []interface{}{
			"id", item.ID,
			"type", item.Type,
		}
		if name := extractName(item); name != nil {
			logFields = append(logFields, "name", *name)
		}
		logFields = append(logFields, "vector_size", len(result.Embedding))

		slog.InfoContext(ctx, "Processed content item", logFields...)
	}

	if len(results) > 0 {
		if err := p.store.StoreAll(ctx, results); err != nil {
			return fmt.Errorf("failed to store results: %w", err)
		}
		slog.InfoContext(ctx, "Stored embedding results", "count", len(results))
	}

	return nil
}

// ProcessItem processes a single ContentItem.
func (p *GenericPipeline) ProcessItem(ctx context.Context, item *ContentItem) (*EmbeddingResult, error) {
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
