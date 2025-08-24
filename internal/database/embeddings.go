package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
)

// EmbeddingVector represents a vector for semantic similarity search.
type EmbeddingVector []float32

// ToText converts the embedding vector to a text representation.
func (ev EmbeddingVector) ToText() string {
	if len(ev) == 0 {
		return "[]"
	}

	result := "["
	for i, v := range ev {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%g", v)
	}
	result += "]"
	return result
}

// ParseEmbeddingVector parses a text representation into an embedding vector.
func ParseEmbeddingVector(text string) (EmbeddingVector, error) {
	if text == "" || text == "[]" {
		return EmbeddingVector{}, nil
	}

	// Simple parsing - in production you might want more robust parsing
	if len(text) < 3 || text[0] != '[' || text[len(text)-1] != ']' {
		return nil, fmt.Errorf("invalid vector format")
	}

	// For now, return empty vector - implement proper parsing as needed
	return EmbeddingVector{}, nil
}

// SimilaritySearchResult represents a similarity search result.
type SimilaritySearchResult struct {
	ID       pgtype.UUID `json:"id"`
	Distance float64     `json:"distance"`
	Name     string      `json:"name"`
}

// SearchSpellsBySimilarity performs similarity search on spells using vector embeddings.
func (d *Database) SearchSpellsBySimilarity(ctx context.Context, embedding EmbeddingVector, limit int32) ([]SearchSpellsByEmbeddingRow, error) {
	// Convert to pgvector.Vector
	pgVector := pgvector.NewVector(embedding)

	return d.queries.SearchSpellsByEmbedding(ctx, SearchSpellsByEmbeddingParams{
		Column1: pgVector,
		Limit:   limit,
	})
}

// SearchCreaturesBySimilarity performs similarity search on creatures using vector embeddings.
func (d *Database) SearchCreaturesBySimilarity(ctx context.Context, embedding EmbeddingVector, limit int32) ([]SearchBeastsByEmbeddingRow, error) {
	// Convert to pgvector.Vector
	pgVector := pgvector.NewVector(embedding)

	return d.queries.SearchBeastsByEmbedding(ctx, SearchBeastsByEmbeddingParams{
		Column1: pgVector,
		Limit:   limit,
	})
}

// SearchClassesBySimilarity performs similarity search on classes using vector embeddings.
func (d *Database) SearchClassesBySimilarity(ctx context.Context, embedding EmbeddingVector, limit int32) ([]SearchClassesByEmbeddingRow, error) {
	// Convert to pgvector.Vector
	pgVector := pgvector.NewVector(embedding)

	return d.queries.SearchClassesByEmbedding(ctx, SearchClassesByEmbeddingParams{
		Column1: pgVector,
		Limit:   limit,
	})
}

// SearchSpeciesBySimilarity performs similarity search on species using vector embeddings.
func (d *Database) SearchSpeciesBySimilarity(ctx context.Context, embedding EmbeddingVector, limit int32) ([]SearchSpeciesByEmbeddingRow, error) {
	// Convert to pgvector.Vector
	pgVector := pgvector.NewVector(embedding)

	return d.queries.SearchSpeciesByEmbedding(ctx, SearchSpeciesByEmbeddingParams{
		Column1: pgVector,
		Limit:   limit,
	})
}
