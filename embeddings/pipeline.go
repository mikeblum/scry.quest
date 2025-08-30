package embeddings //nolint:revive // package comment not needed

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/pgvector/pgvector-go"
)

// Pipeline processes SRD content for embedding generation
type Pipeline struct {
	client  *Client
	queries *database.Queries
	srdPath string
}

// NewPipeline creates a new content processing pipeline
func NewPipeline(client *Client, queries *database.Queries, srdPath string) *Pipeline {
	return &Pipeline{
		client:  client,
		queries: queries,
		srdPath: srdPath,
	}
}

// ProcessAll processes all content types
func (p *Pipeline) ProcessAll(ctx context.Context) error {
	contentTypes := []ContentType{
		ContentTypeSpell,
		ContentTypeBestiary,
		ContentTypeClass,
		ContentTypeSpecies,
	}

	for _, ct := range contentTypes {
		if err := p.ProcessContentType(ctx, ct); err != nil {
			return fmt.Errorf("failed to process %s: %w", ct, err)
		}
	}

	return nil
}

// ProcessContentType processes a specific content type
func (p *Pipeline) ProcessContentType(ctx context.Context, contentType ContentType) error {
	slog.InfoContext(ctx, "Processing content type", "type", contentType)

	switch contentType {
	case ContentTypeSpell:
		return p.processSpells(ctx)
	case ContentTypeBestiary:
		return p.processBestiary(ctx)
	case ContentTypeClass:
		return p.processClasses(ctx)
	case ContentTypeSpecies:
		return p.processSpecies(ctx)
	default:
		return fmt.Errorf("unknown content type: %s", contentType)
	}
}

func (p *Pipeline) processSpells(ctx context.Context) error {
	spellsPath := filepath.Join(p.srdPath, "spells", "spells.md")
	content, err := os.ReadFile(spellsPath) //nolint:gosec // filepath.Join ensures safe path construction
	if err != nil {
		return fmt.Errorf("failed to read spells file: %w", err)
	}

	text := string(content)
	embedding, err := p.client.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	vector := pgvector.NewVector(embedding)
	modelText := pgtype.Text{String: p.client.config.Model, Valid: true}

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
		RawData:        content,
		EmbeddingModel: modelText,
	}

	if _, err := p.queries.CreateSpell(ctx, params); err != nil {
		return fmt.Errorf("failed to create spell record: %w", err)
	}

	slog.InfoContext(ctx, "Processed spells", "embedding_dims", len(embedding))
	return nil
}

func (p *Pipeline) processBestiary(ctx context.Context) error {
	bestiaryPath := filepath.Join(p.srdPath, "beastiary")
	entries, err := os.ReadDir(bestiaryPath)
	if err != nil {
		return fmt.Errorf("failed to read bestiary directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(bestiaryPath, entry.Name())
		if err := p.processBestiaryFile(ctx, filePath); err != nil {
			slog.ErrorContext(ctx, "Failed to process bestiary file", "file", filePath, "error", err)
		}
	}

	return nil
}

func (p *Pipeline) processBestiaryFile(ctx context.Context, filePath string) error {
	content, err := os.ReadFile(filePath) //nolint:gosec // filepath.Join ensures safe path construction
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var creature map[string]interface{}
	if err := json.Unmarshal(content, &creature); err != nil {
		return fmt.Errorf("failed to unmarshal creature data: %w", err)
	}

	params, err := p.buildCreatureParams(ctx, creature, content)
	if err != nil {
		return err
	}

	if _, err := p.queries.CreateCreature(ctx, *params); err != nil {
		return fmt.Errorf("failed to create creature record: %w", err)
	}

	slog.InfoContext(ctx, "Processed creature", "name", params.Name)
	return nil
}

func (p *Pipeline) buildCreatureParams(ctx context.Context, creature map[string]interface{}, content []byte) (*database.CreateCreatureParams, error) {
	name, ok := creature["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid name field")
	}

	text := fmt.Sprintf("Name: %s\n%s", name, string(content))
	embedding, err := p.client.GenerateEmbedding(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	vector := pgvector.NewVector(embedding)
	modelText := pgtype.Text{String: p.client.config.Model, Valid: true}

	abilities, _ := json.Marshal(creature["ability_scores"])
	skills, _ := json.Marshal(creature["skills"])
	speed, _ := json.Marshal(creature["speed"])

	params := &database.CreateCreatureParams{
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
		RawData:        content,
		EmbeddingModel: modelText,
	}

	p.setOptionalCreatureFields(creature, params)
	return params, nil
}

func (p *Pipeline) setOptionalCreatureFields(creature map[string]interface{}, params *database.CreateCreatureParams) {
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
}

func (p *Pipeline) processClasses(ctx context.Context) error {
	classesPath := filepath.Join(p.srdPath, "classes")
	entries, err := os.ReadDir(classesPath)
	if err != nil {
		return fmt.Errorf("failed to read classes directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(classesPath, entry.Name())
		if err := p.processClassFile(ctx, filePath); err != nil {
			slog.ErrorContext(ctx, "Failed to process class file", "file", filePath, "error", err)
		}
	}

	return nil
}

func (p *Pipeline) processClassFile(ctx context.Context, filePath string) error {
	content, err := os.ReadFile(filePath) //nolint:gosec // filepath.Join ensures safe path construction
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fileName := filepath.Base(filePath)
	className := toTitleCase(strings.TrimSuffix(fileName, ".md"))

	text := string(content)
	embedding, err := p.client.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	vector := pgvector.NewVector(embedding)
	modelText := pgtype.Text{String: p.client.config.Model, Valid: true}

	params := database.CreateClassParams{
		Name:           className,
		Description:    pgtype.Text{String: text[:minInt(1000, len(text))], Valid: true},
		Embedding:      vector,
		RawData:        content,
		EmbeddingModel: modelText,
	}

	if _, err := p.queries.CreateClass(ctx, params); err != nil {
		return fmt.Errorf("failed to create class record: %w", err)
	}

	slog.InfoContext(ctx, "Processed class", "name", className, "embedding_dims", len(embedding))
	return nil
}

func (p *Pipeline) processSpecies(ctx context.Context) error {
	speciesPath := filepath.Join(p.srdPath, "species")
	entries, err := os.ReadDir(speciesPath)
	if err != nil {
		return fmt.Errorf("failed to read species directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		filePath := filepath.Join(speciesPath, entry.Name())
		if err := p.processSpeciesFile(ctx, filePath); err != nil {
			slog.ErrorContext(ctx, "Failed to process species file", "file", filePath, "error", err)
		}
	}

	return nil
}

func (p *Pipeline) processSpeciesFile(ctx context.Context, filePath string) error {
	content, err := os.ReadFile(filePath) //nolint:gosec // filepath.Join ensures safe path construction
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fileName := filepath.Base(filePath)
	speciesName := toTitleCase(strings.TrimSuffix(fileName, ".md"))

	text := string(content)
	embedding, err := p.client.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	vector := pgvector.NewVector(embedding)
	modelText := pgtype.Text{String: p.client.config.Model, Valid: true}

	params := database.CreateSpeciesParams{
		Name:           speciesName,
		Description:    pgtype.Text{String: text[:minInt(1000, len(text))], Valid: true},
		Traits:         []string{},
		Embedding:      vector,
		RawData:        content,
		EmbeddingModel: modelText,
	}

	if _, err := p.queries.CreateSpecies(ctx, params); err != nil {
		return fmt.Errorf("failed to create species record: %w", err)
	}

	slog.InfoContext(ctx, "Processed species", "name", speciesName, "embedding_dims", len(embedding))
	return nil
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func toTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}
