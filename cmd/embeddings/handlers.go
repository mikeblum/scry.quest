package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mikeblum/scry.quest/conf"
	"github.com/mikeblum/scry.quest/embeddings"
	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/urfave/cli/v2"
)

func generateEmbeddings(c *cli.Context, engine *embeddings.Engine) error {
	// Create SRD-specific pipeline
	config := embeddings.SRDPipelineConfig{
		SRDPath: "./srd",
		Client:  engine.Client,
		Queries: engine.Queries,
		Conn:    engine.Conn,
		Model:   embeddings.Model(engine.Config.GetPrefixedEnv("OLLAMA_MODEL", "nomic-embed-text")),
	}
	pipeline := embeddings.CreateSRDPipeline(config)

	contentType := c.String("type")
	if contentType == "" {
		slog.Info("Generating embeddings for all content types...")
		return embeddings.ProcessAllSRDContent(c.Context, pipeline, config.SRDPath)
	}

	contentTypeMap := map[string]embeddings.ContentType{
		"spell":    embeddings.ContentTypeSpell,
		"bestiary": embeddings.ContentTypeBestiary,
		"class":    embeddings.ContentTypeClass,
		"species":  embeddings.ContentTypeSpecies,
	}

	if ct, ok := contentTypeMap[contentType]; ok {
		return embeddings.ProcessSRDContentType(c.Context, pipeline, ct, config.SRDPath)
	}

	return fmt.Errorf("unknown content type: %s", contentType)
}

func searchContent(c *cli.Context, engine *embeddings.Engine) error {
	searchService := embeddings.NewSearchService(engine.Client, engine.Queries)
	query := c.String("query")
	contentType := c.String("type")
	limit := c.Int("limit")

	var results []*embeddings.SearchResult
	var err error

	if contentType == "" {
		results, err = searchAllContentTypes(c.Context, searchService, query, limit)
	} else {
		results, err = searchSpecificContentType(c.Context, searchService, query, contentType, limit)
	}

	if err != nil {
		return err
	}

	displaySearchResults(results, query)
	return nil
}

func searchAllContentTypes(ctx context.Context, searchService *embeddings.SearchService, query string, limit int) ([]*embeddings.SearchResult, error) {
	if limit < 0 || limit > 1000 {
		limit = 10
	}
	limit32 := int32(limit) //nolint:gosec // bounded by check above

	return searchService.Search(ctx, query, &embeddings.SearchOptions{
		ContentTypes: []embeddings.ContentType{
			embeddings.ContentTypeSpell,
			embeddings.ContentTypeBestiary,
			embeddings.ContentTypeClass,
			embeddings.ContentTypeSpecies,
		},
		Limit:     limit32,
		Threshold: 0.6,
	})
}

func searchSpecificContentType(ctx context.Context, searchService *embeddings.SearchService, query, contentType string, limit int) ([]*embeddings.SearchResult, error) {
	if limit < 0 || limit > 1000 {
		limit = 10
	}
	limit32 := int32(limit) //nolint:gosec // bounded by check above

	switch contentType {
	case "spell":
		return searchService.SearchSpells(ctx, query, limit32)
	case "bestiary":
		return searchService.SearchBestiary(ctx, query, limit32)
	case "class":
		return searchService.SearchClasses(ctx, query, limit32)
	case "species":
		return searchService.SearchSpecies(ctx, query, limit32)
	default:
		return nil, fmt.Errorf("unknown content type: %s", contentType)
	}
}

func displaySearchResults(results []*embeddings.SearchResult, query string) {
	slog.Info("Search results", "count", len(results), "query", query)

	for i, result := range results {
		content := ""
		if result.Content != "" && len(result.Content) > 100 {
			content = result.Content[:100] + "..."
		} else if result.Content != "" {
			content = result.Content
		}
		slog.Info("Result", "rank", i+1, "name", result.Name, "type", result.Type, "similarity", result.Similarity, "content", content)
	}
}

func showStats(c *cli.Context, engine *embeddings.Engine) error {
	stats, err := engine.Queries.GetEmbeddingStats(c.Context)
	if err != nil {
		return err
	}

	counts, err := engine.Queries.CountItemsByEmbeddingModel(c.Context)
	if err != nil {
		return err
	}

	report := generateStatsReport(stats, counts)
	slog.Info("Embedding statistics", "report", report)

	return nil
}

func clearEmbeddings(c *cli.Context, engine *embeddings.Engine) error {
	ollamaModel := engine.Config.String("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = string(embeddings.Embedding) // defaults to "nomic-embed-text"
	}
	if c.String("ollama-model") != "" {
		ollamaModel = c.String("ollama-model")
	}

	slog.Info("Clearing embeddings for model", "model", ollamaModel)

	modelText := pgtype.Text{String: ollamaModel, Valid: true}

	if err := engine.Queries.DeleteSpellEmbeddings(c.Context, modelText); err != nil {
		return fmt.Errorf("failed to clear spell embeddings: %w", err)
	}

	if err := engine.Queries.DeleteCreatureEmbeddings(c.Context, modelText); err != nil {
		return fmt.Errorf("failed to clear creature embeddings: %w", err)
	}

	if err := engine.Queries.DeleteClassEmbeddings(c.Context, modelText); err != nil {
		return fmt.Errorf("failed to clear class embeddings: %w", err)
	}

	if err := engine.Queries.DeleteSpeciesEmbeddings(c.Context, modelText); err != nil {
		return fmt.Errorf("failed to clear species embeddings: %w", err)
	}

	slog.Info("Successfully cleared embeddings for model", "model", ollamaModel)
	return nil
}

func runMigrations(c *cli.Context) error {
	configPath := c.String("config")
	config, err := conf.New(c.Context, &configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	databaseURL := config.GetPrefixedEnv("DATABASE_URL", "postgres://localhost/scry_quest?sslmode=disable")

	dbConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	db, err := database.NewDatabase(c.Context, database.Config{
		Host:     dbConfig.Host,
		Port:     fmt.Sprintf("%d", dbConfig.Port),
		User:     dbConfig.User,
		Password: dbConfig.Password,
		Database: dbConfig.Database,
		SSLMode:  "disable",
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func(c *cli.Context) {
		_ = db.Close(c.Context)
	}(c)

	if err := db.RunMigrations(c.Context, database.Config{
		Host:     dbConfig.Host,
		Port:     fmt.Sprintf("%d", dbConfig.Port),
		User:     dbConfig.User,
		Password: dbConfig.Password,
		Database: dbConfig.Database,
		SSLMode:  "disable",
	}); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Database migrations completed successfully")
	return nil
}

func generateStatsReport(_ interface{}, _ interface{}) string {
	return "Embedding statistics - implementation needed"
}
