package embeddings

import (
	"context"
	"fmt"

	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/pgvector/pgvector-go"
)

// SearchResult represents a search result with similarity score
type SearchResult struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       ContentType            `json:"type"`
	Content    string                 `json:"content,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Similarity float64                `json:"similarity"`
}

// SearchService provides semantic search functionality
type SearchService struct {
	client *Client
	db     *database.Queries
}

// NewSearchService creates a new search service
func NewSearchService(client *Client, db *database.Queries) *SearchService {
	return &SearchService{
		client: client,
		db:     db,
	}
}

// SearchOptions configures search behavior
type SearchOptions struct {
	ContentTypes []ContentType `json:"content_types,omitempty"`
	Limit        int32         `json:"limit,omitempty"`
	Threshold    float64       `json:"threshold,omitempty"`
}

// DefaultSearchOptions returns sensible default search options
func DefaultSearchOptions() *SearchOptions {
	return &SearchOptions{
		ContentTypes: []ContentType{
			ContentTypeSpell,
			ContentTypeBestiary,
			ContentTypeClass,
			ContentTypeSpecies,
		},
		Limit:     10,
		Threshold: 0.7, // Minimum similarity score
	}
}

// Search performs semantic search across all content types
func (s *SearchService) Search(ctx context.Context, query string, options *SearchOptions) ([]*SearchResult, error) {
	if options == nil {
		options = DefaultSearchOptions()
	}

	// Generate embedding for the search query
	queryEmbedding, err := s.client.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	var allResults []*SearchResult

	// Search in each content type
	for _, contentType := range options.ContentTypes {
		results, err := s.searchContentType(ctx, contentType, queryEmbedding, options)
		if err != nil {
			return nil, fmt.Errorf("failed to search %s content: %w", contentType, err)
		}
		allResults = append(allResults, results...)
	}

	// Sort by similarity and apply limit
	allResults = s.sortAndLimitResults(allResults, options.Limit)

	return allResults, nil
}

// SearchSpells performs semantic search specifically for spells
func (s *SearchService) SearchSpells(ctx context.Context, query string, limit int32) ([]*SearchResult, error) {
	queryEmbedding, err := s.client.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return s.searchContentType(ctx, ContentTypeSpell, queryEmbedding, &SearchOptions{
		Limit:     limit,
		Threshold: 0.7,
	})
}

// SearchBestiary performs semantic search specifically for creatures
func (s *SearchService) SearchBestiary(ctx context.Context, query string, limit int32) ([]*SearchResult, error) {
	queryEmbedding, err := s.client.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return s.searchContentType(ctx, ContentTypeBestiary, queryEmbedding, &SearchOptions{
		Limit:     limit,
		Threshold: 0.7,
	})
}

// SearchClasses performs semantic search specifically for classes
func (s *SearchService) SearchClasses(ctx context.Context, query string, limit int32) ([]*SearchResult, error) {
	queryEmbedding, err := s.client.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return s.searchContentType(ctx, ContentTypeClass, queryEmbedding, &SearchOptions{
		Limit:     limit,
		Threshold: 0.7,
	})
}

// SearchSpecies performs semantic search specifically for species
func (s *SearchService) SearchSpecies(ctx context.Context, query string, limit int32) ([]*SearchResult, error) {
	queryEmbedding, err := s.client.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return s.searchContentType(ctx, ContentTypeSpecies, queryEmbedding, &SearchOptions{
		Limit:     limit,
		Threshold: 0.7,
	})
}

// searchContentType performs similarity search for a specific content type
func (s *SearchService) searchContentType(ctx context.Context, contentType ContentType, queryEmbedding []float32, options *SearchOptions) ([]*SearchResult, error) {
	embedding := pgvector.NewVector(queryEmbedding)

	switch contentType {
	case ContentTypeSpell:
		return s.searchSpellsDB(ctx, embedding, options)
	case ContentTypeBestiary:
		return s.searchBestiaryDB(ctx, embedding, options)
	case ContentTypeClass:
		return s.searchClassesDB(ctx, embedding, options)
	case ContentTypeSpecies:
		return s.searchSpeciesDB(ctx, embedding, options)
	default:
		return nil, fmt.Errorf("unknown content type: %s", contentType)
	}
}

// searchSpellsDB searches spells using vector similarity
func (s *SearchService) searchSpellsDB(ctx context.Context, queryEmbedding pgvector.Vector, options *SearchOptions) ([]*SearchResult, error) {
	params := database.SearchSpellsByEmbeddingParams{
		Embedding: queryEmbedding,
		Limit:     options.Limit,
	}

	rows, err := s.db.SearchSpellsByEmbedding(ctx, params)
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, 0, len(rows))
	for _, row := range rows {
		if float64(row.Similarity) < options.Threshold {
			continue
		}

		result := &SearchResult{
			ID:         row.ID.String(),
			Name:       row.Name,
			Type:       ContentTypeSpell,
			Similarity: float64(row.Similarity),
		}

		if row.Description.Valid {
			result.Content = row.Description.String
		}

		results = append(results, result)
	}

	return results, nil
}

// searchBestiaryDB searches bestiary using vector similarity
func (s *SearchService) searchBestiaryDB(ctx context.Context, queryEmbedding pgvector.Vector, options *SearchOptions) ([]*SearchResult, error) {
	params := database.SearchCreaturesByEmbeddingParams{
		Embedding: queryEmbedding,
		Limit:     options.Limit,
	}

	rows, err := s.db.SearchCreaturesByEmbedding(ctx, params)
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, 0, len(rows))
	for _, row := range rows {
		if float64(row.Similarity) < options.Threshold {
			continue
		}

		result := &SearchResult{
			ID:         row.ID.String(),
			Name:       row.Name,
			Type:       ContentTypeBestiary,
			Similarity: float64(row.Similarity),
		}

		results = append(results, result)
	}

	return results, nil
}

// searchClassesDB searches classes using vector similarity
func (s *SearchService) searchClassesDB(ctx context.Context, queryEmbedding pgvector.Vector, options *SearchOptions) ([]*SearchResult, error) {
	params := database.SearchClassesByEmbeddingParams{
		Embedding: queryEmbedding,
		Limit:     options.Limit,
	}

	rows, err := s.db.SearchClassesByEmbedding(ctx, params)
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, 0, len(rows))
	for _, row := range rows {
		if float64(row.Similarity) < options.Threshold {
			continue
		}

		result := &SearchResult{
			ID:         row.ID.String(),
			Name:       row.Name,
			Type:       ContentTypeClass,
			Similarity: float64(row.Similarity),
		}

		if row.Description.Valid {
			result.Content = row.Description.String
		}

		results = append(results, result)
	}

	return results, nil
}

// searchSpeciesDB searches species using vector similarity
func (s *SearchService) searchSpeciesDB(ctx context.Context, queryEmbedding pgvector.Vector, options *SearchOptions) ([]*SearchResult, error) {
	params := database.SearchSpeciesByEmbeddingParams{
		Embedding: queryEmbedding,
		Limit:     options.Limit,
	}

	rows, err := s.db.SearchSpeciesByEmbedding(ctx, params)
	if err != nil {
		return nil, err
	}

	results := make([]*SearchResult, 0, len(rows))
	for _, row := range rows {
		if float64(row.Similarity) < options.Threshold {
			continue
		}

		result := &SearchResult{
			ID:         row.ID.String(),
			Name:       row.Name,
			Type:       ContentTypeSpecies,
			Similarity: float64(row.Similarity),
		}

		if row.Description.Valid {
			result.Content = row.Description.String
		}

		results = append(results, result)
	}

	return results, nil
}

// sortAndLimitResults sorts results by similarity and applies limit
func (s *SearchService) sortAndLimitResults(results []*SearchResult, limit int32) []*SearchResult {
	// Sort by similarity descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && int(limit) < len(results) {
		results = results[:limit]
	}

	return results
}
