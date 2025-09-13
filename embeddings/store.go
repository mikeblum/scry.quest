package embeddings

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/pgvector/pgvector-go"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// DatabaseEmbeddingStore stores embeddings in PostgreSQL.
type DatabaseEmbeddingStore struct {
	queries *database.Queries
	conn    *pgx.Conn
}

// NewDatabaseEmbeddingStore creates a database store with connection for transactions.
func NewDatabaseEmbeddingStore(queries *database.Queries, conn *pgx.Conn) *DatabaseEmbeddingStore {
	return &DatabaseEmbeddingStore{
		queries: queries,
		conn:    conn,
	}
}

// Store implements the EmbeddingStore interface
func (s *DatabaseEmbeddingStore) Store(ctx context.Context, result *EmbeddingResult) error {
	vector := pgvector.NewVector(result.Embedding)
	model := getString(result.Metadata, "model", "nomic-embed-text")
	modelText := pgtype.Text{String: model, Valid: true}

	originalContent := result.Metadata["original_content"].([]byte)
	description := string(originalContent)
	if len(description) > 1000 {
		description = description[:1000]
	}

	// Convert content to valid JSON for storage
	rawData, err := ensureValidJSON(originalContent)
	if err != nil {
		return fmt.Errorf("failed to convert content to JSON: %w", err)
	}

	// Extract a simple name from the content or use filename
	name := extractSimpleName(result, string(originalContent))
	if name == "" {
		name = result.ContentID.String() // Use UUID as fallback
	}

	switch result.ContentType {
	case string(ContentTypeSpell):
		params := database.CreateSpellParams{
			Name:           name,
			Description:    pgtype.Text{String: description, Valid: true},
			Level:          0, // Default level, actual parsing can be added later if needed
			School:         pgtype.Text{String: "N/A", Valid: true},
			CastingTime:    pgtype.Text{String: "N/A", Valid: true},
			RangeValue:     pgtype.Text{String: "N/A", Valid: true},
			Components:     pgtype.Text{String: "N/A", Valid: true},
			Duration:       pgtype.Text{String: "N/A", Valid: true},
			Classes:        []string{},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		}
		_, err := s.queries.CreateSpell(ctx, params)
		return err

	case string(ContentTypeBestiary):
		params := database.CreateCreatureParams{
			Name:           name,
			Size:           pgtype.Text{String: "Medium", Valid: true}, // Default
			Type:           pgtype.Text{Valid: false},
			Alignment:      pgtype.Text{String: "Unaligned", Valid: true},
			Abilities:      []byte("{}"),
			Skills:         []byte("{}"),
			Speed:          []byte("{}"),
			Languages:      pgtype.Text{String: "—", Valid: true},
			Senses:         pgtype.Text{String: "—", Valid: true},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		}
		_, err := s.queries.CreateCreature(ctx, params)
		return err

	case string(ContentTypeClass):
		params := database.CreateClassParams{
			Name:           name,
			Description:    pgtype.Text{String: description, Valid: true},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		}
		_, err := s.queries.CreateClass(ctx, params)
		return err

	case string(ContentTypeSpecies):
		params := database.CreateSpeciesParams{
			Name:           name,
			Description:    pgtype.Text{String: description, Valid: true},
			Traits:         []string{},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		}
		_, err := s.queries.CreateSpecies(ctx, params)
		return err

	default:
		return fmt.Errorf("unsupported content type: %s", result.ContentType)
	}
}

// StoreAll implements the EmbeddingStore interface with batch transaction support
func (s *DatabaseEmbeddingStore) StoreAll(ctx context.Context, results []*EmbeddingResult) error {
	return s.batchInsertTx(ctx, results)
}

// batchInsertTx performs efficient batch insertion using sqlc within a transaction
func (s *DatabaseEmbeddingStore) batchInsertTx(ctx context.Context, results []*EmbeddingResult) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txQueries := s.queries.WithTx(tx)

	if err := s.batchInsertByType(ctx, txQueries, results); err != nil {
		return fmt.Errorf("failed to batch insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// batchInsertByType groups results by content type and performs batch inserts
func (s *DatabaseEmbeddingStore) batchInsertByType(ctx context.Context, queries *database.Queries, results []*EmbeddingResult) error {
	grouped := s.groupResultsByType(results)

	if err := s.insertSpells(ctx, queries, grouped.spells); err != nil {
		return err
	}
	if err := s.insertCreatures(ctx, queries, grouped.creatures); err != nil {
		return err
	}
	if err := s.insertClasses(ctx, queries, grouped.classes); err != nil {
		return err
	}
	if err := s.insertSpecies(ctx, queries, grouped.species); err != nil {
		return err
	}

	return nil
}

type groupedResults struct {
	spells    []database.CreateSpellsParams
	creatures []database.CreateCreaturesParams
	classes   []database.CreateClassesParams
	species   []database.CreateSpeciesBatchParams
}

func (s *DatabaseEmbeddingStore) groupResultsByType(results []*EmbeddingResult) groupedResults {
	var grouped groupedResults

	for _, result := range results {
		params := s.createParamsFromResult(result)

		switch result.ContentType {
		case string(ContentTypeSpell):
			grouped.spells = append(grouped.spells, params.spell)
		case string(ContentTypeBestiary):
			grouped.creatures = append(grouped.creatures, params.creature)
		case string(ContentTypeClass):
			grouped.classes = append(grouped.classes, params.class)
		case string(ContentTypeSpecies):
			grouped.species = append(grouped.species, params.species)
		}
	}

	return grouped
}

type batchParams struct {
	spell    database.CreateSpellsParams
	creature database.CreateCreaturesParams
	class    database.CreateClassesParams
	species  database.CreateSpeciesBatchParams
}

func (s *DatabaseEmbeddingStore) createParamsFromResult(result *EmbeddingResult) batchParams {
	vector := pgvector.NewVector(result.Embedding)
	model := getString(result.Metadata, "model", "nomic-embed-text")
	modelText := pgtype.Text{String: model, Valid: true}

	originalContent := result.Metadata["original_content"].([]byte)
	description := string(originalContent)
	if len(description) > 1000 {
		description = description[:1000]
	}

	rawData, _ := ensureValidJSON(originalContent)
	name := extractSimpleName(result, string(originalContent))
	if name == "" {
		name = result.ContentID.String()
	}

	return batchParams{
		spell: database.CreateSpellsParams{
			Name:           name,
			Description:    pgtype.Text{String: description, Valid: true},
			Level:          0,
			School:         pgtype.Text{String: "N/A", Valid: true},
			CastingTime:    pgtype.Text{String: "N/A", Valid: true},
			RangeValue:     pgtype.Text{String: "N/A", Valid: true},
			Components:     pgtype.Text{String: "N/A", Valid: true},
			Duration:       pgtype.Text{String: "N/A", Valid: true},
			Classes:        []string{},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		},
		creature: database.CreateCreaturesParams{
			Name:           name,
			Size:           pgtype.Text{String: "Medium", Valid: true},
			Type:           pgtype.Text{Valid: false},
			Alignment:      pgtype.Text{String: "Unaligned", Valid: true},
			Abilities:      []byte("{}"),
			Skills:         []byte("{}"),
			Speed:          []byte("{}"),
			Languages:      pgtype.Text{String: "—", Valid: true},
			Senses:         pgtype.Text{String: "—", Valid: true},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		},
		class: database.CreateClassesParams{
			Name:           name,
			Description:    pgtype.Text{String: description, Valid: true},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		},
		species: database.CreateSpeciesBatchParams{
			Name:           name,
			Description:    pgtype.Text{String: description, Valid: true},
			Traits:         []string{},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		},
	}
}

func (s *DatabaseEmbeddingStore) insertSpells(ctx context.Context, queries *database.Queries, params []database.CreateSpellsParams) error {
	if len(params) > 0 {
		br := queries.CreateSpells(ctx, params)
		var batchErr error
		br.Exec(func(i int, err error) {
			if err != nil && batchErr == nil {
				batchErr = fmt.Errorf("failed to insert spell at index %d: %w", i, err)
			}
		})
		if batchErr != nil {
			return batchErr
		}
	}
	return nil
}

func (s *DatabaseEmbeddingStore) insertCreatures(ctx context.Context, queries *database.Queries, params []database.CreateCreaturesParams) error {
	if len(params) > 0 {
		br := queries.CreateCreatures(ctx, params)
		var batchErr error
		br.Exec(func(i int, err error) {
			if err != nil && batchErr == nil {
				batchErr = fmt.Errorf("failed to insert creature at index %d: %w", i, err)
			}
		})
		if batchErr != nil {
			return batchErr
		}
	}
	return nil
}

func (s *DatabaseEmbeddingStore) insertClasses(ctx context.Context, queries *database.Queries, params []database.CreateClassesParams) error {
	if len(params) > 0 {
		br := queries.CreateClasses(ctx, params)
		var batchErr error
		br.Exec(func(i int, err error) {
			if err != nil && batchErr == nil {
				batchErr = fmt.Errorf("failed to insert class at index %d: %w", i, err)
			}
		})
		if batchErr != nil {
			return batchErr
		}
	}
	return nil
}

func (s *DatabaseEmbeddingStore) insertSpecies(ctx context.Context, queries *database.Queries, params []database.CreateSpeciesBatchParams) error {
	if len(params) > 0 {
		br := queries.CreateSpeciesBatch(ctx, params)
		var batchErr error
		br.Exec(func(i int, err error) {
			if err != nil && batchErr == nil {
				batchErr = fmt.Errorf("failed to insert species at index %d: %w", i, err)
			}
		})
		if batchErr != nil {
			return batchErr
		}
	}
	return nil
}

// Simple helpers
func getString(metadata map[string]any, key, defaultValue string) string {
	if val, ok := metadata[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func extractSimpleName(result *EmbeddingResult, content string) string {
	if name := extractFromFilename(result); name != "" {
		return name
	}
	if name := extractFromJSON(content); name != "" {
		return name
	}
	if name := extractFromMarkdown(content); name != "" {
		return name
	}
	return ""
}

func extractFromFilename(result *EmbeddingResult) string {
	filename, ok := result.Metadata["filename"].(string)
	if !ok {
		return ""
	}
	name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	name = strings.ReplaceAll(name, "_", " ")
	return cases.Title(language.English).String(strings.ToLower(name))
}

func extractFromJSON(content string) string {
	if !strings.HasPrefix(content, "{") {
		return ""
	}
	lines := strings.Split(content, "\n")
	limit := min(5, len(lines))
	for _, line := range lines[:limit] {
		if name := parseJSONNameLine(line); name != "" {
			return name
		}
	}
	return ""
}

func parseJSONNameLine(line string) string {
	if !strings.Contains(line, `"name"`) || !strings.Contains(line, ":") {
		return ""
	}
	parts := strings.Split(line, ":")
	if len(parts) <= 1 {
		return ""
	}
	name := strings.Trim(strings.Trim(parts[1], `,"`), `"`)
	return name
}

func extractFromMarkdown(content string) string {
	if !strings.HasPrefix(content, "#") {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.TrimLeft(lines[0], "#"))
}

// ensureValidJSON converts content to valid JSON for database storage.
// If content is already valid JSON, returns it as-is.
// If not, wraps it in a JSON object with a "content" field.
func ensureValidJSON(content []byte) ([]byte, error) {
	// First check if it's already valid JSON
	var dummy interface{}
	if err := json.Unmarshal(content, &dummy); err == nil {
		return content, nil
	}

	// If not valid JSON, wrap it in a JSON object
	wrapper := map[string]string{
		"content": string(content),
	}
	return json.Marshal(wrapper)
}
