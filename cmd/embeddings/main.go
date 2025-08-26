package main //nolint:revive // package comment not needed for main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mikeblum/scry.quest/conf"
	"github.com/mikeblum/scry.quest/embeddings"
	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/mikeblum/scry.quest/log"
)

const (
	defaultOllamaHost  = "http://localhost:11434"
	defaultOllamaModel = "gpt-oss:20b"
)

type config struct {
	configPath  *string
	command     *string
	model       *string
	ollamaURL   *string
	contentType *string
	query       *string
	limit       *int
	srdPath     *string
}

func parseFlags() *config {
	cfg := &config{
		configPath:  flag.String("config", "env/.env.local", "Path to configuration file"),
		command:     flag.String("command", "", "Command to run: generate, search, stats, clear"),
		model:       flag.String("model", defaultOllamaModel, "Embedding model to use (gpt-oss:20b, nomic-embed-text, all-minilm)"),
		ollamaURL:   flag.String("ollama-url", defaultOllamaHost, "Ollama server URL"),
		contentType: flag.String("type", "", "Content type filter (spell, bestiary, class, species)"),
		query:       flag.String("query", "", "Search query (for search command)"),
		limit:       flag.Int("limit", 10, "Number of results to return (for search command)"),
		srdPath:     flag.String("srd", "./srd", "Path to SRD directory"),
	}
	flag.Parse()
	return cfg
}

func showUsage() {
	slog.Info("Usage: embeddings -command <generate|search|stats|clear> [options]")
	slog.Info("")
	slog.Info("Commands:")
	slog.Info("  generate    Generate embeddings for SRD content")
	slog.Info("  search      Search content using embeddings")
	slog.Info("  stats       Show embedding statistics")
	slog.Info("  clear       Clear embeddings for a model")
	slog.Info("")
	slog.Info("Options:")
	flag.PrintDefaults()
}

func setupServices(ctx context.Context, cfg *config) (*embeddings.Client, *database.Queries, func() error, error) {
	// Load configuration
	appCfg, err := conf.New(ctx, *cfg.configPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database connection
	conn, err := pgx.Connect(ctx, appCfg.String("DATABASE_URL"))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	cleanup := func() error {
		return conn.Close(ctx)
	}

	queries := database.New(conn)

	// Initialize Ollama client
	ollamaConfig := embeddings.Config{
		Host:  *cfg.ollamaURL,
		Model: *cfg.model,
	}
	client, err := embeddings.NewClient(ollamaConfig)
	if err != nil {
		if cleanupErr := cleanup(); cleanupErr != nil {
			slog.Error("Failed to cleanup during client creation failure", "error", cleanupErr)
		}
		return nil, nil, nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	// Test Ollama connection
	if err := client.Ping(ctx); err != nil {
		if cleanupErr := cleanup(); cleanupErr != nil {
			slog.Error("Failed to cleanup during ping failure", "error", cleanupErr)
		}
		return nil, nil, nil, fmt.Errorf("failed to connect to Ollama server at %s: %w", *cfg.ollamaURL, err)
	}

	slog.Info("Connected to Ollama server", "url", *cfg.ollamaURL, "model", *cfg.model)
	return client, queries, cleanup, nil
}

func runCommand(ctx context.Context, cfg *config, client *embeddings.Client, queries *database.Queries) error {
	switch *cfg.command {
	case "generate":
		return generateEmbeddings(ctx, client, queries, *cfg.srdPath, *cfg.contentType)
	case "search":
		if *cfg.query == "" {
			return fmt.Errorf("query is required for search command")
		}
		return searchContent(ctx, client, queries, *cfg.query, *cfg.contentType, *cfg.limit)
	case "stats":
		return showStats(ctx, queries)
	case "clear":
		return clearEmbeddings(ctx, queries, *cfg.model)
	default:
		return fmt.Errorf("unknown command: %s", *cfg.command)
	}
}

func main() {
	cfg := parseFlags()

	if *cfg.command == "" {
		showUsage()
		return
	}

	// Initialize logging
	log.NewFromEnv()

	ctx := context.Background()

	client, queries, cleanup, err := setupServices(ctx, cfg)
	if err != nil {
		slog.Error("Failed to setup services", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := cleanup(); err != nil {
			slog.Error("Failed to cleanup", "error", err)
		}
	}()

	if err := runCommand(ctx, cfg, client, queries); err != nil {
		slog.Error("Command failed", "error", err)
		return
	}
}

func generateEmbeddings(ctx context.Context, client *embeddings.Client, queries *database.Queries, srdPath, contentType string) error {
	pipeline := embeddings.NewPipeline(client, queries, srdPath)

	if contentType == "" {
		slog.Info("Generating embeddings for all content types...")
		return pipeline.ProcessAll(ctx)
	}

	contentTypeMap := map[string]embeddings.ContentType{
		"spell":    embeddings.ContentTypeSpell,
		"bestiary": embeddings.ContentTypeBestiary,
		"class":    embeddings.ContentTypeClass,
		"species":  embeddings.ContentTypeSpecies,
	}

	if ct, ok := contentTypeMap[contentType]; ok {
		return pipeline.ProcessContentType(ctx, ct)
	}

	return fmt.Errorf("unknown content type: %s", contentType)
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

func searchContent(ctx context.Context, client *embeddings.Client, queries *database.Queries, query, contentType string, limit int) error {
	searchService := embeddings.NewSearchService(client, queries)

	var results []*embeddings.SearchResult
	var err error

	if contentType == "" {
		results, err = searchAllContentTypes(ctx, searchService, query, limit)
	} else {
		results, err = searchSpecificContentType(ctx, searchService, query, contentType, limit)
	}

	if err != nil {
		return err
	}

	displaySearchResults(results, query)
	return nil
}

func showStats(ctx context.Context, queries *database.Queries) error {
	stats, err := queries.GetEmbeddingStats(ctx)
	if err != nil {
		return err
	}

	slog.Info("Embedding Statistics")

	for _, stat := range stats {
		expectedDims := "N/A"
		if stat.ExpectedDimensions != nil {
			if dims, ok := stat.ExpectedDimensions.(int32); ok {
				expectedDims = fmt.Sprintf("%d", dims)
			}
		}

		embeddingModel := "none"
		if stat.EmbeddingModel.Valid {
			embeddingModel = stat.EmbeddingModel.String
		}

		slog.Info("Table stats", "table", stat.TableName, "model", embeddingModel, "total", stat.TotalRows, "embedded", stat.EmbeddedRows, "expected_dims", expectedDims)
	}

	// Show counts by model
	counts, err := queries.CountItemsByEmbeddingModel(ctx)
	if err != nil {
		return err
	}

	slog.Info("Counts by Model")
	for _, count := range counts {
		slog.Info("Model count", "table", count.TableName, "model", count.Model, "count", count.Count)
	}

	return nil
}

func clearEmbeddings(ctx context.Context, queries *database.Queries, model string) error {
	slog.Info("Clearing embeddings for model", "model", model)

	modelText := pgtype.Text{String: model, Valid: true}

	if err := queries.DeleteSpellEmbeddings(ctx, modelText); err != nil {
		return fmt.Errorf("failed to clear spell embeddings: %w", err)
	}

	if err := queries.DeleteCreatureEmbeddings(ctx, modelText); err != nil {
		return fmt.Errorf("failed to clear creature embeddings: %w", err)
	}

	if err := queries.DeleteClassEmbeddings(ctx, modelText); err != nil {
		return fmt.Errorf("failed to clear class embeddings: %w", err)
	}

	if err := queries.DeleteSpeciesEmbeddings(ctx, modelText); err != nil {
		return fmt.Errorf("failed to clear species embeddings: %w", err)
	}

	slog.Info("Successfully cleared embeddings for model", "model", model)
	return nil
}
