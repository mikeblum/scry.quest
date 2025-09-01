package embeddings

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/pgvector/pgvector-go"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// DatabaseEmbeddingStore stores embeddings in PostgreSQL.
type DatabaseEmbeddingStore struct {
	queries *database.Queries
}

// NewDatabaseEmbeddingStore creates a database store.
func NewDatabaseEmbeddingStore(queries *database.Queries) *DatabaseEmbeddingStore {
	return &DatabaseEmbeddingStore{
		queries: queries,
	}
}

// Store stores an embedding result.
func (s *DatabaseEmbeddingStore) Store(ctx context.Context, result *EmbeddingResult) error {
	switch result.ContentType {
	case string(ContentTypeSpell):
		return s.storeSpell(ctx, result)
	case string(ContentTypeBestiary):
		return s.storeCreature(ctx, result)
	case string(ContentTypeClass):
		return s.storeClass(ctx, result)
	case string(ContentTypeSpecies):
		return s.storeSpecies(ctx, result)
	default:
		return fmt.Errorf("unsupported content type: %s", result.ContentType)
	}
}

// StoreAll stores multiple results.
func (s *DatabaseEmbeddingStore) StoreAll(ctx context.Context, results []*EmbeddingResult) error {
	for _, result := range results {
		if err := s.Store(ctx, result); err != nil {
			return fmt.Errorf("failed to store result %s: %w", result.ContentID, err)
		}
	}
	return nil
}

func (s *DatabaseEmbeddingStore) storeSpell(ctx context.Context, result *EmbeddingResult) error {
	vector := pgvector.NewVector(result.Embedding)
	model := getStringFromMetadata(result.Metadata, "model", nil)
	modelText := pgtype.Text{String: model, Valid: true}

	// For spells, we typically have one large document
	params := database.CreateSpellParams{
		Name:           "D&D 5e Spells",
		Description:    pgtype.Text{String: "Complete D&D 5e spell reference", Valid: true},
		Level:          0,
		School:         pgtype.Text{String: "reference", Valid: true},
		CastingTime:    pgtype.Text{String: "N/A", Valid: true},
		RangeValue:     pgtype.Text{String: "N/A", Valid: true},
		Components:     pgtype.Text{String: "N/A", Valid: true},
		Duration:       pgtype.Text{String: "N/A", Valid: true},
		Classes:        []string{"reference"},
		Embedding:      vector,
		RawData:        result.Metadata["original_content"].([]byte),
		EmbeddingModel: modelText,
	}

	if _, err := s.queries.CreateSpell(ctx, params); err != nil {
		return fmt.Errorf("failed to create spell record: %w", err)
	}

	return nil
}

func (s *DatabaseEmbeddingStore) storeCreature(ctx context.Context, result *EmbeddingResult) error {
	vector := pgvector.NewVector(result.Embedding)
	model := getStringFromMetadata(result.Metadata, "model", nil)
	modelText := pgtype.Text{String: model, Valid: true}

	// Extract creature data from metadata
	var creature map[string]interface{}
	if jsonData, ok := result.Metadata["json_data"].(map[string]interface{}); ok {
		creature = jsonData
	} else {
		return fmt.Errorf("missing creature JSON data in metadata")
	}

	contentIDStr := result.ContentID.String()
	name := getStringFromMetadata(creature, "name", &contentIDStr)

	abilities, _ := json.Marshal(creature["ability_scores"])
	skills, _ := json.Marshal(creature["skills"])
	speed, _ := json.Marshal(creature["speed"])

	params := database.CreateCreatureParams{
		Name:           name,
		Size:           getTextFromMap(creature, "size"),
		Type:           getTextFromMap(creature, "type"),
		Alignment:      getTextFromMap(creature, "alignment"),
		Abilities:      abilities,
		Skills:         skills,
		Speed:          speed,
		Languages:      getTextFromMap(creature, "languages"),
		Senses:         getTextFromMap(creature, "senses"),
		Embedding:      vector,
		RawData:        result.Metadata["original_content"].([]byte),
		EmbeddingModel: modelText,
	}

	// Set optional fields
	if armorClass, ok := creature["armor_class"].(float64); ok {
		params.ArmorClass = pgtype.Int4{Int32: int32(armorClass), Valid: true}
	}

	if hitPoints, ok := creature["hit_points"].(map[string]interface{}); ok {
		if avg, ok := hitPoints["average"].(float64); ok {
			params.HitPoints = pgtype.Int4{Int32: int32(avg), Valid: true}
		}
		if dice, ok := hitPoints["dice"].(string); ok {
			params.HitDice = pgtype.Text{String: dice, Valid: true}
		}
	}

	if cr, ok := creature["challenge_rating"].(map[string]interface{}); ok {
		if rating, ok := cr["rating"].(string); ok {
			params.ChallengeRating = pgtype.Text{String: rating, Valid: true}
		}
	}

	if _, err := s.queries.CreateCreature(ctx, params); err != nil {
		return fmt.Errorf("failed to create creature record: %w", err)
	}

	return nil
}

func (s *DatabaseEmbeddingStore) storeClass(ctx context.Context, result *EmbeddingResult) error {
	vector := pgvector.NewVector(result.Embedding)
	model := getStringFromMetadata(result.Metadata, "model", nil)
	modelText := pgtype.Text{String: model, Valid: true}

	contentIDStr := result.ContentID.String()
	name := getStringFromMetadata(result.Metadata, "base_name", &contentIDStr)
	name = cases.Title(language.English).String(strings.ToLower(name))

	content := string(result.Metadata["content"].([]byte))
	description := content
	if len(content) > 1000 {
		description = content[:1000]
	}

	params := database.CreateClassParams{
		Name:           name,
		Description:    pgtype.Text{String: description, Valid: true},
		Embedding:      vector,
		RawData:        result.Metadata["original_content"].([]byte),
		EmbeddingModel: modelText,
	}

	if _, err := s.queries.CreateClass(ctx, params); err != nil {
		return fmt.Errorf("failed to create class record: %w", err)
	}

	return nil
}

func (s *DatabaseEmbeddingStore) storeSpecies(ctx context.Context, result *EmbeddingResult) error {
	vector := pgvector.NewVector(result.Embedding)
	model := getStringFromMetadata(result.Metadata, "model", nil)
	modelText := pgtype.Text{String: model, Valid: true}

	contentIDStr := result.ContentID.String()
	name := getStringFromMetadata(result.Metadata, "base_name", &contentIDStr)
	name = cases.Title(language.English).String(strings.ToLower(name))

	content := string(result.Metadata["content"].([]byte))
	description := content
	if len(content) > 1000 {
		description = content[:1000]
	}

	params := database.CreateSpeciesParams{
		Name:           name,
		Description:    pgtype.Text{String: description, Valid: true},
		Traits:         []string{},
		Embedding:      vector,
		RawData:        result.Metadata["original_content"].([]byte),
		EmbeddingModel: modelText,
	}

	if _, err := s.queries.CreateSpecies(ctx, params); err != nil {
		return fmt.Errorf("failed to create species record: %w", err)
	}

	return nil
}

// Helpers
func getStringFromMetadata(metadata map[string]interface{}, key string, defaultValue *string) string {
	if val, ok := metadata[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	if defaultValue != nil {
		return *defaultValue
	}
	return ""
}

func getTextFromMap(m map[string]interface{}, key string) pgtype.Text {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return pgtype.Text{String: str, Valid: true}
		}
		if slice, ok := val.([]interface{}); ok && len(slice) > 0 {
			var strs []string
			for _, item := range slice {
				if s, ok := item.(string); ok {
					strs = append(strs, s)
				}
			}
			if len(strs) > 0 {
				return pgtype.Text{String: strings.Join(strs, ", "), Valid: true}
			}
		}
	}
	return pgtype.Text{}
}
