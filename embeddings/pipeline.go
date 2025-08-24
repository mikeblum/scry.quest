package embeddings

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/pgvector/pgvector-go"
)

// ContentType represents the type of SRD content
type ContentType string

const (
	ContentTypeSpell    ContentType = "spell"
	ContentTypeBestiary ContentType = "bestiary"
	ContentTypeClass    ContentType = "class"
	ContentTypeSpecies  ContentType = "species"
)

// ContentItem represents a processed piece of SRD content
type ContentItem struct {
	Type      ContentType
	Name      string
	Content   string
	FilePath  string
	Metadata  map[string]interface{}
	Embedding []float32
}

// Pipeline processes SRD content and generates embeddings
type Pipeline struct {
	client    *Client
	db        *database.Queries
	srdPath   string
	batchSize int
}

// NewPipeline creates a new embeddings pipeline
func NewPipeline(client *Client, db *database.Queries, srdPath string) *Pipeline {
	return &Pipeline{
		client:    client,
		db:        db,
		srdPath:   srdPath,
		batchSize: 10, // Process in batches to avoid overwhelming the API
	}
}

// ProcessAll processes all SRD content and stores embeddings
func (p *Pipeline) ProcessAll(ctx context.Context) error {
	contentTypes := []ContentType{
		ContentTypeSpell,
		ContentTypeBestiary,
		ContentTypeClass,
		ContentTypeSpecies,
	}

	for _, contentType := range contentTypes {
		fmt.Printf("Processing %s content...\n", contentType)
		if err := p.ProcessContentType(ctx, contentType); err != nil {
			return fmt.Errorf("failed to process %s content: %w", contentType, err)
		}
	}

	return nil
}

// ProcessContentType processes all content of a specific type
func (p *Pipeline) ProcessContentType(ctx context.Context, contentType ContentType) error {
	items, err := p.loadContentItems(contentType)
	if err != nil {
		return fmt.Errorf("failed to load content items: %w", err)
	}

	fmt.Printf("Found %d %s items\n", len(items), contentType)

	// Process in batches
	for i := 0; i < len(items); i += p.batchSize {
		end := i + p.batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]
		if err := p.processBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to process batch %d-%d: %w", i, end-1, err)
		}

		fmt.Printf("Processed %d/%d %s items\n", end, len(items), contentType)

		// Small delay to be respectful to the Ollama server
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// loadContentItems loads content items from the SRD directory
func (p *Pipeline) loadContentItems(contentType ContentType) ([]*ContentItem, error) {
	var items []*ContentItem
	var searchPath string
	var fileExt string

	switch contentType {
	case ContentTypeSpell:
		searchPath = filepath.Join(p.srdPath, "spells")
		fileExt = ".json"
	case ContentTypeBestiary:
		searchPath = filepath.Join(p.srdPath, "beastiary")
		fileExt = ".json"
	case ContentTypeClass:
		searchPath = filepath.Join(p.srdPath, "classes")
		fileExt = ".md"
	case ContentTypeSpecies:
		searchPath = filepath.Join(p.srdPath, "species")
		fileExt = ".md"
	default:
		return nil, fmt.Errorf("unknown content type: %s", contentType)
	}

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, fileExt) {
			item, err := p.loadContentItem(contentType, path)
			if err != nil {
				fmt.Printf("Warning: failed to load %s: %v\n", path, err)
				return nil // Continue processing other files
			}
			items = append(items, item)
		}

		return nil
	})

	return items, err
}

// loadContentItem loads a single content item from file
func (p *Pipeline) loadContentItem(contentType ContentType, filePath string) (*ContentItem, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	item := &ContentItem{
		Type:     contentType,
		FilePath: filePath,
		Metadata: make(map[string]interface{}),
	}

	switch contentType {
	case ContentTypeSpell, ContentTypeBestiary:
		var jsonData map[string]interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}

		// Extract name
		if name, ok := jsonData["name"].(string); ok {
			item.Name = name
		} else {
			return nil, fmt.Errorf("missing or invalid name field")
		}

		// Store full JSON as metadata
		item.Metadata = jsonData

		// Create searchable text content
		item.Content = p.createSearchableContent(contentType, jsonData)

	case ContentTypeClass, ContentTypeSpecies:
		content := string(data)
		item.Content = content

		// Extract name from markdown content (first heading)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "# ") {
				item.Name = strings.TrimPrefix(line, "# ")
				break
			}
		}

		if item.Name == "" {
			// Fallback to filename without extension
			item.Name = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		}

		item.Metadata["file_path"] = filePath
		item.Metadata["type"] = string(contentType)
	}

	return item, nil
}

// createSearchableContent creates searchable text from structured data
func (p *Pipeline) createSearchableContent(contentType ContentType, data map[string]interface{}) string {
	var parts []string

	switch contentType {
	case ContentTypeSpell:
		if name, ok := data["name"].(string); ok {
			parts = append(parts, name)
		}
		if school, ok := data["school"].(string); ok {
			parts = append(parts, school)
		}
		if level, ok := data["level"].(string); ok {
			parts = append(parts, level)
		}
		if desc, ok := data["description"].(string); ok {
			parts = append(parts, desc)
		}
		if higher, ok := data["at_higher_levels"].(string); ok {
			parts = append(parts, higher)
		}

	case ContentTypeBestiary:
		if name, ok := data["name"].(string); ok {
			parts = append(parts, name)
		}
		if size, ok := data["size"].(string); ok {
			parts = append(parts, size)
		}
		if creatureType, ok := data["type"].(string); ok {
			parts = append(parts, creatureType)
		}
		if alignment, ok := data["alignment"].(string); ok {
			parts = append(parts, alignment)
		}

		// Add trait descriptions
		if traits, ok := data["traits"].([]interface{}); ok {
			for _, trait := range traits {
				if traitMap, ok := trait.(map[string]interface{}); ok {
					if traitName, ok := traitMap["name"].(string); ok {
						parts = append(parts, traitName)
					}
					if traitDesc, ok := traitMap["description"].(string); ok {
						parts = append(parts, traitDesc)
					}
				}
			}
		}
	}

	return strings.Join(parts, " ")
}

// processBatch processes a batch of content items
func (p *Pipeline) processBatch(ctx context.Context, items []*ContentItem) error {
	// Extract texts for batch embedding generation
	texts := make([]string, len(items))
	for i, item := range items {
		texts[i] = item.Content
	}

	// Generate embeddings in batch
	embeddings, err := p.client.GenerateEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Store embeddings and content in database
	for i, item := range items {
		item.Embedding = embeddings[i]
		if err := p.storeItem(ctx, item); err != nil {
			return fmt.Errorf("failed to store item %s: %w", item.Name, err)
		}
	}

	return nil
}

// storeItem stores a content item with its embedding in the database
func (p *Pipeline) storeItem(ctx context.Context, item *ContentItem) error {
	embedding := pgvector.NewVector(item.Embedding)

	switch item.Type {
	case ContentTypeSpell:
		return p.storeSpell(ctx, item, embedding)
	case ContentTypeBestiary:
		return p.storeBeastiary(ctx, item, embedding)
	case ContentTypeClass:
		return p.storeClass(ctx, item, embedding)
	case ContentTypeSpecies:
		return p.storeSpecies(ctx, item, embedding)
	default:
		return fmt.Errorf("unknown content type: %s", item.Type)
	}
}

// storeSpell stores a spell in the database
func (p *Pipeline) storeSpell(ctx context.Context, item *ContentItem, embedding pgvector.Vector) error {
	rawData, _ := json.Marshal(item.Metadata)

	params := database.CreateSpellParams{
		Name:        item.Name,
		Description: pgtype.Text{String: item.Content, Valid: true},
		Embedding:   embedding,
		RawData:     rawData,
	}

	// Extract specific fields from metadata
	if level, ok := item.Metadata["level"].(string); ok {
		// Parse level from "Level X" format
		if strings.HasPrefix(level, "Level ") {
			levelNum := strings.TrimPrefix(level, "Level ")
			if len(levelNum) > 0 && levelNum[0] >= '0' && levelNum[0] <= '9' {
				params.Level = int32(levelNum[0] - '0')
			}
		}
	}

	if school, ok := item.Metadata["school"].(string); ok {
		params.School = pgtype.Text{String: school, Valid: true}
	}

	if castingTime, ok := item.Metadata["casting_time"].(map[string]interface{}); ok {
		if value, ok := castingTime["value"].(string); ok {
			if unit, ok := castingTime["unit"].(string); ok {
				castingTimeStr := fmt.Sprintf("%s %s", value, unit)
				params.CastingTime = pgtype.Text{String: castingTimeStr, Valid: true}
			}
		}
	}

	_, err := p.db.CreateSpell(ctx, params)
	return err
}

// storeBeastiary stores a creature in the bestiary table
func (p *Pipeline) storeBeastiary(ctx context.Context, item *ContentItem, embedding pgvector.Vector) error {
	rawData, _ := json.Marshal(item.Metadata)

	params := database.CreateCreatureParams{
		Name:      item.Name,
		Embedding: embedding,
		RawData:   rawData,
	}

	// Extract specific fields from metadata
	if size, ok := item.Metadata["size"].(string); ok {
		params.Size = pgtype.Text{String: size, Valid: true}
	}

	if creatureType, ok := item.Metadata["type"].(string); ok {
		params.Type = pgtype.Text{String: creatureType, Valid: true}
	}

	if alignment, ok := item.Metadata["alignment"].(string); ok {
		params.Alignment = pgtype.Text{String: alignment, Valid: true}
	}

	if ac, ok := item.Metadata["armor_class"].(float64); ok {
		acInt := int32(ac)
		params.ArmorClass = pgtype.Int4{Int32: acInt, Valid: true}
	}

	if hp, ok := item.Metadata["hit_points"].(map[string]interface{}); ok {
		if average, ok := hp["average"].(float64); ok {
			hpInt := int32(average)
			params.HitPoints = pgtype.Int4{Int32: hpInt, Valid: true}
		}
		if dice, ok := hp["dice"].(string); ok {
			params.HitDice = pgtype.Text{String: dice, Valid: true}
		}
	}

	if cr, ok := item.Metadata["challenge_rating"].(map[string]interface{}); ok {
		if rating, ok := cr["rating"].(string); ok {
			params.ChallengeRating = pgtype.Text{String: rating, Valid: true}
		}
	}

	_, err := p.db.CreateCreature(ctx, params)
	return err
}

// storeClass stores a class in the database
func (p *Pipeline) storeClass(ctx context.Context, item *ContentItem, embedding pgvector.Vector) error {
	rawData, _ := json.Marshal(item.Metadata)

	params := database.CreateClassParams{
		Name:        item.Name,
		Description: pgtype.Text{String: item.Content, Valid: true},
		Embedding:   embedding,
		RawData:     rawData,
	}

	_, err := p.db.CreateClass(ctx, params)
	return err
}

// storeSpecies stores a species in the database
func (p *Pipeline) storeSpecies(ctx context.Context, item *ContentItem, embedding pgvector.Vector) error {
	rawData, _ := json.Marshal(item.Metadata)

	params := database.CreateSpeciesParams{
		Name:        item.Name,
		Description: pgtype.Text{String: item.Content, Valid: true},
		Embedding:   embedding,
		RawData:     rawData,
	}

	_, err := p.db.CreateSpecies(ctx, params)
	return err
}
