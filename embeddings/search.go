package embeddings //nolint:revive // package comment not needed

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/pgvector/pgvector-go"
)

// SearchService provides embedding-based search functionality
type SearchService struct {
	client  *Client
	queries *database.Queries
}

// NewSearchService creates a new search service
func NewSearchService(client *Client, queries *database.Queries) *SearchService {
	return &SearchService{
		client:  client,
		queries: queries,
	}
}

// Search performs semantic search across multiple content types
func (s *SearchService) Search(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if opts == nil {
		opts = &SearchOptions{
			ContentTypes: []ContentType{ContentTypeSpell, ContentTypeBestiary, ContentTypeClass, ContentTypeSpecies},
			Limit:        10,
			Threshold:    0.6,
		}
	}

	var allResults []*SearchResult

	for _, contentType := range opts.ContentTypes {
		results, err := s.searchByType(ctx, query, contentType, opts.Limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search %s: %w", contentType, err)
		}
		allResults = append(allResults, results...)
	}

	return s.filterAndLimitResults(allResults, opts), nil
}

func (s *SearchService) searchByType(ctx context.Context, query string, contentType ContentType, limit int32) ([]*SearchResult, error) {
	embedding, err := s.client.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	vector := pgvector.NewVector(embedding)

	switch contentType {
	case ContentTypeSpell:
		return s.searchSpells(ctx, vector, limit)
	case ContentTypeBestiary:
		return s.searchBestiary(ctx, vector, limit)
	case ContentTypeClass:
		return s.searchClasses(ctx, vector, limit)
	case ContentTypeSpecies:
		return s.searchSpecies(ctx, vector, limit)
	default:
		return nil, fmt.Errorf("unknown content type: %s", contentType)
	}
}

func (s *SearchService) searchSpells(ctx context.Context, vector pgvector.Vector, limit int32) ([]*SearchResult, error) {
	spells, err := s.queries.SearchSpellsByEmbedding(ctx, database.SearchSpellsByEmbeddingParams{
		Embedding: vector,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, len(spells))
	for i, spell := range spells {
		results[i] = &SearchResult{
			ID:         uuid.UUID(spell.ID.Bytes).String(),
			Name:       spell.Name,
			Type:       ContentTypeSpell,
			Content:    spell.Description.String,
			Similarity: float64(spell.Similarity),
		}
	}
	return results, nil
}

func (s *SearchService) searchBestiary(ctx context.Context, vector pgvector.Vector, limit int32) ([]*SearchResult, error) {
	creatures, err := s.queries.SearchCreaturesByEmbedding(ctx, database.SearchCreaturesByEmbeddingParams{
		Embedding: vector,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, len(creatures))
	for i, creature := range creatures {
		results[i] = &SearchResult{
			ID:         uuid.UUID(creature.ID.Bytes).String(),
			Name:       creature.Name,
			Type:       ContentTypeBestiary,
			Content:    fmt.Sprintf("%s %s", creature.Type.String, creature.Size.String),
			Similarity: float64(creature.Similarity),
		}
	}
	return results, nil
}

func (s *SearchService) searchClasses(ctx context.Context, vector pgvector.Vector, limit int32) ([]*SearchResult, error) {
	classes, err := s.queries.SearchClassesByEmbedding(ctx, database.SearchClassesByEmbeddingParams{
		Embedding: vector,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, len(classes))
	for i, class := range classes {
		results[i] = &SearchResult{
			ID:         uuid.UUID(class.ID.Bytes).String(),
			Name:       class.Name,
			Type:       ContentTypeClass,
			Content:    class.Description.String,
			Similarity: float64(class.Similarity),
		}
	}
	return results, nil
}

func (s *SearchService) searchSpecies(ctx context.Context, vector pgvector.Vector, limit int32) ([]*SearchResult, error) {
	species, err := s.queries.SearchSpeciesByEmbedding(ctx, database.SearchSpeciesByEmbeddingParams{
		Embedding: vector,
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, len(species))
	for i, sp := range species {
		results[i] = &SearchResult{
			ID:         uuid.UUID(sp.ID.Bytes).String(),
			Name:       sp.Name,
			Type:       ContentTypeSpecies,
			Content:    sp.Description.String,
			Similarity: float64(sp.Similarity),
		}
	}
	return results, nil
}

// SearchSpells searches spell content
func (s *SearchService) SearchSpells(ctx context.Context, query string, limit int32) ([]*SearchResult, error) {
	return s.searchByType(ctx, query, ContentTypeSpell, limit)
}

// SearchBestiary searches bestiary content
func (s *SearchService) SearchBestiary(ctx context.Context, query string, limit int32) ([]*SearchResult, error) {
	return s.searchByType(ctx, query, ContentTypeBestiary, limit)
}

// SearchClasses searches class content
func (s *SearchService) SearchClasses(ctx context.Context, query string, limit int32) ([]*SearchResult, error) {
	return s.searchByType(ctx, query, ContentTypeClass, limit)
}

// SearchSpecies searches species content
func (s *SearchService) SearchSpecies(ctx context.Context, query string, limit int32) ([]*SearchResult, error) {
	return s.searchByType(ctx, query, ContentTypeSpecies, limit)
}

func (s *SearchService) filterAndLimitResults(results []*SearchResult, opts *SearchOptions) []*SearchResult {
	var filtered []*SearchResult

	for _, result := range results {
		if result.Similarity >= opts.Threshold {
			filtered = append(filtered, result)
		}
	}

	if len(filtered) > int(opts.Limit) {
		filtered = filtered[:opts.Limit]
	}

	return filtered
}
