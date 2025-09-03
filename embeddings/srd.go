package embeddings

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/mikeblum/scry.quest/internal/database"
)

// SRDPipelineConfig holds SRD processing configuration.
type SRDPipelineConfig struct {
	SRDPath string
	Client  *Client
	Queries *database.Queries
	Model   Model
}

// CreateSRDPipeline creates an SRD content pipeline.
func CreateSRDPipeline(config SRDPipelineConfig) *Pipeline {
	processor := NewDefaultContentProcessor(config.Client, config.Model)
	store := NewDatabaseEmbeddingStore(config.Queries)

	return NewPipeline(config.Client, processor, store)
}

// ProcessSRDContentType processes specific SRD content type.
func ProcessSRDContentType(ctx context.Context, pipeline *Pipeline, contentType ContentType, srdPath string) error {
	var source DataSource
	var err error

	switch contentType {
	case ContentTypeSpell:
		source, err = createSpellDataSource(srdPath)
	case ContentTypeBestiary:
		source, err = createBestiaryDataSource(srdPath)
	case ContentTypeClass:
		source, err = createClassDataSource(srdPath)
	case ContentTypeSpecies:
		source, err = createSpeciesDataSource(srdPath)
	default:
		return fmt.Errorf("unknown content type: %s", contentType)
	}

	if err != nil {
		return fmt.Errorf("failed to create data source for %s: %w", contentType, err)
	}

	slog.InfoContext(ctx, "Processing SRD content type", "type", contentType)
	return pipeline.ProcessSource(ctx, source)
}

// ProcessAllSRDContent processes all SRD content types.
func ProcessAllSRDContent(ctx context.Context, pipeline *Pipeline, srdPath string) error {
	contentTypes := []ContentType{
		ContentTypeSpell,
		ContentTypeBestiary,
		ContentTypeClass,
		ContentTypeSpecies,
	}

	for _, ct := range contentTypes {
		if err := ProcessSRDContentType(ctx, pipeline, ct, srdPath); err != nil {
			return fmt.Errorf("failed to process %s: %w", ct, err)
		}
	}

	return nil
}

// Data sources
func createSpellDataSource(srdPath string) (DataSource, error) {
	spellsPath := filepath.Join(srdPath, "spells")
	return NewFileSystemDataSource(spellsPath, string(ContentTypeSpell), []string{".md"}), nil
}

func createBestiaryDataSource(srdPath string) (DataSource, error) {
	bestiaryPath := filepath.Join(srdPath, "beastiary") // Note: keeping original typo for compatibility
	return NewFileSystemDataSource(bestiaryPath, string(ContentTypeBestiary), []string{".json"}), nil
}

func createClassDataSource(srdPath string) (DataSource, error) {
	classesPath := filepath.Join(srdPath, "classes")
	return NewFileSystemDataSource(classesPath, string(ContentTypeClass), []string{".md"}), nil
}

func createSpeciesDataSource(srdPath string) (DataSource, error) {
	speciesPath := filepath.Join(srdPath, "species")
	return NewFileSystemDataSource(speciesPath, string(ContentTypeSpecies), []string{".md"}), nil
}
