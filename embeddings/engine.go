package embeddings //nolint:revive // package comment not needed

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mikeblum/scry.quest/conf"
	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/mikeblum/scry.quest/log"
	"github.com/urfave/cli/v2"
)

// Engine represents the embeddings processing engine
type Engine struct {
	config *conf.Config
}

// NewEngine creates a new embeddings processing engine
func NewEngine() *cli.App {
	engine := &Engine{}

	return &cli.App{
		Name:  "embeddings",
		Usage: "D&D 5e SRD embeddings pipeline",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "Path to configuration file",
			},
			&cli.StringFlag{
				Name:    "model",
				Usage:   "Override embedding model",
				EnvVars: []string{"OLLAMA_MODEL"},
			},
			&cli.StringFlag{
				Name:    "ollama-url",
				Usage:   "Override Ollama server URL",
				EnvVars: []string{"OLLAMA_HOST"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "Generate embeddings for SRD content",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "type",
						Usage: "Content type filter (spell, bestiary, class, species)",
					},
				},
				Action: func(c *cli.Context) error {
					return engine.runWithSetup(c, engine.generateEmbeddings)
				},
			},
			{
				Name:  "search",
				Usage: "Search content using embeddings",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "query",
						Usage:    "Search query",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "type",
						Usage: "Content type filter (spell, bestiary, class, species)",
					},
					&cli.IntFlag{
						Name:  "limit",
						Usage: "Number of results to return",
						Value: 10,
					},
				},
				Action: func(c *cli.Context) error {
					return engine.runWithSetup(c, engine.searchContent)
				},
			},
			{
				Name:  "stats",
				Usage: "Show embedding statistics",
				Action: func(c *cli.Context) error {
					return engine.runWithSetup(c, engine.showStats)
				},
			},
			{
				Name:  "clear",
				Usage: "Clear embeddings for a model",
				Action: func(c *cli.Context) error {
					return engine.runWithSetup(c, engine.clearEmbeddings)
				},
			},
		},
	}
}

func (e *Engine) loadConfig(c *cli.Context) error {
	// Load configuration using conf module
	config, err := conf.New(c.Context, c.String("config"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	e.config = config
	return nil
}

func (e *Engine) runWithSetup(c *cli.Context, handler func(*cli.Context, *Client, *database.Queries) error) error {
	// Load configuration first
	if err := e.loadConfig(c); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logging using the log module's environment configuration
	log.NewFromEnv()

	client, queries, cleanup, err := e.setupServices(c.Context, c)
	if err != nil {
		return fmt.Errorf("failed to setup services: %w", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			slog.Error("Failed to cleanup", "error", err)
		}
	}()

	return handler(c, client, queries)
}

func (e *Engine) setupServices(ctx context.Context, c *cli.Context) (*Client, *database.Queries, func() error, error) {
	databaseURL := e.getDatabaseURL()
	ollamaHost := e.getOllamaHost(c)
	ollamaModel := e.getOllamaModel(c)

	// Initialize database connection
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	cleanup := func() error {
		return conn.Close(ctx)
	}

	queries := database.New(conn)

	// Initialize and test Ollama client
	client, err := e.createAndTestOllamaClient(ctx, ollamaHost, ollamaModel, cleanup)
	if err != nil {
		return nil, nil, nil, err
	}

	slog.Info("Connected to Ollama server", "url", ollamaHost, "model", ollamaModel)
	return client, queries, cleanup, nil
}

func (e *Engine) generateEmbeddings(c *cli.Context, client *Client, queries *database.Queries) error {
	pipeline := NewPipeline(client, queries, "./srd")

	contentType := c.String("type")
	if contentType == "" {
		slog.Info("Generating embeddings for all content types...")
		return pipeline.ProcessAll(c.Context)
	}

	contentTypeMap := map[string]ContentType{
		"spell":    ContentTypeSpell,
		"bestiary": ContentTypeBestiary,
		"class":    ContentTypeClass,
		"species":  ContentTypeSpecies,
	}

	if ct, ok := contentTypeMap[contentType]; ok {
		return pipeline.ProcessContentType(c.Context, ct)
	}

	return fmt.Errorf("unknown content type: %s", contentType)
}

func (e *Engine) searchContent(c *cli.Context, client *Client, queries *database.Queries) error {
	searchService := NewSearchService(client, queries)
	query := c.String("query")
	contentType := c.String("type")
	limit := c.Int("limit")

	var results []*SearchResult
	var err error

	if contentType == "" {
		results, err = e.searchAllContentTypes(c.Context, searchService, query, limit)
	} else {
		results, err = e.searchSpecificContentType(c.Context, searchService, query, contentType, limit)
	}

	if err != nil {
		return err
	}

	e.displaySearchResults(results, query)
	return nil
}

func (e *Engine) searchAllContentTypes(ctx context.Context, searchService *SearchService, query string, limit int) ([]*SearchResult, error) {
	if limit < 0 || limit > 1000 {
		limit = 10
	}
	limit32 := int32(limit) //nolint:gosec // bounded by check above

	return searchService.Search(ctx, query, &SearchOptions{
		ContentTypes: []ContentType{
			ContentTypeSpell,
			ContentTypeBestiary,
			ContentTypeClass,
			ContentTypeSpecies,
		},
		Limit:     limit32,
		Threshold: 0.6,
	})
}

func (e *Engine) searchSpecificContentType(ctx context.Context, searchService *SearchService, query, contentType string, limit int) ([]*SearchResult, error) {
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

func (e *Engine) displaySearchResults(results []*SearchResult, query string) {
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

func (e *Engine) showStats(c *cli.Context, _ *Client, queries *database.Queries) error {
	stats, err := queries.GetEmbeddingStats(c.Context)
	if err != nil {
		return err
	}

	counts, err := queries.CountItemsByEmbeddingModel(c.Context)
	if err != nil {
		return err
	}

	report := e.generateStatsReport(stats, counts)
	slog.Info("Embedding statistics", "report", report)

	return nil
}

func (e *Engine) generateStatsReport(stats []database.ScryQuestEmbeddingStat, counts []database.CountItemsByEmbeddingModelRow) string {
	var report strings.Builder

	report.WriteString("# D&D 5e SRD Embeddings Statistics\n\n")

	e.writeTableOverview(&report, stats)
	modelCounts := e.writeModelBreakdown(&report, counts)
	e.writeSummary(&report, stats, modelCounts)

	return report.String()
}

func (e *Engine) writeTableOverview(report *strings.Builder, stats []database.ScryQuestEmbeddingStat) {
	report.WriteString("## Table Overview\n\n")
	report.WriteString("| Table | Model | Total Rows | Embedded | Coverage | Dimensions |\n")
	report.WriteString("|-------|-------|------------|----------|----------|------------|\n")

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

		coverage := "0%"
		if stat.TotalRows > 0 {
			percentage := float64(stat.EmbeddedRows) / float64(stat.TotalRows) * 100
			coverage = fmt.Sprintf("%.1f%%", percentage)
		}

		fmt.Fprintf(report, "| %s | %s | %d | %d | %s | %s |\n",
			stat.TableName,
			embeddingModel,
			stat.TotalRows,
			stat.EmbeddedRows,
			coverage,
			expectedDims)
	}
}

func (e *Engine) writeModelBreakdown(report *strings.Builder, counts []database.CountItemsByEmbeddingModelRow) map[string][]database.CountItemsByEmbeddingModelRow {
	report.WriteString("\n## Embeddings by Model\n\n")

	// Group counts by model
	modelCounts := make(map[string][]database.CountItemsByEmbeddingModelRow)
	for _, count := range counts {
		modelCounts[count.Model] = append(modelCounts[count.Model], count)
	}

	for model, modelData := range modelCounts {
		if model == "no_embedding" {
			continue // Skip unembedded items for now
		}

		fmt.Fprintf(report, "### %s\n\n", model)
		report.WriteString("| Table | Count |\n")
		report.WriteString("|-------|-------|\n")

		totalForModel := int64(0)
		for _, data := range modelData {
			fmt.Fprintf(report, "| %s | %d |\n", data.TableName, data.Count)
			totalForModel += data.Count
		}

		fmt.Fprintf(report, "| **Total** | **%d** |\n\n", totalForModel)
	}

	return modelCounts
}

func (e *Engine) writeSummary(report *strings.Builder, stats []database.ScryQuestEmbeddingStat, modelCounts map[string][]database.CountItemsByEmbeddingModelRow) {
	totalEmbedded := int64(0)
	totalRows := int64(0)
	for _, stat := range stats {
		totalEmbedded += stat.EmbeddedRows
		totalRows += stat.TotalRows
	}

	report.WriteString("## Summary\n\n")
	fmt.Fprintf(report, "- **Total Content Items**: %d\n", totalRows)
	fmt.Fprintf(report, "- **Items with Embeddings**: %d\n", totalEmbedded)
	fmt.Fprintf(report, "- **Overall Coverage**: %.1f%%\n", float64(totalEmbedded)/float64(totalRows)*100)
	fmt.Fprintf(report, "- **Models in Use**: %d\n\n", len(modelCounts)-1) // -1 to exclude "no_embedding"
}

func (e *Engine) getDatabaseURL() string {
	databaseURL := e.config.String("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://localhost/scry_quest?sslmode=disable"
	}
	return databaseURL
}

func (e *Engine) getOllamaHost(c *cli.Context) string {
	ollamaHost := e.config.String("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	if c.String("ollama-url") != "" {
		ollamaHost = c.String("ollama-url")
	}
	return ollamaHost
}

func (e *Engine) getOllamaModel(c *cli.Context) string {
	ollamaModel := e.config.String("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "gpt-oss:20b"
	}
	if c.String("model") != "" {
		ollamaModel = c.String("model")
	}
	return ollamaModel
}

func (e *Engine) createAndTestOllamaClient(ctx context.Context, host, model string, cleanup func() error) (*Client, error) {
	ollamaConfig := Config{
		Host:  host,
		Model: model,
	}
	client, err := NewClient(ollamaConfig)
	if err != nil {
		if cleanupErr := cleanup(); cleanupErr != nil {
			slog.Error("Failed to cleanup during client creation failure", "error", cleanupErr)
		}
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	// Test Ollama connection
	if err := client.Ping(ctx); err != nil {
		if cleanupErr := cleanup(); cleanupErr != nil {
			slog.Error("Failed to cleanup during ping failure", "error", cleanupErr)
		}
		return nil, fmt.Errorf("failed to connect to Ollama server at %s: %w", host, err)
	}

	return client, nil
}

func (e *Engine) clearEmbeddings(c *cli.Context, _ *Client, queries *database.Queries) error {
	ollamaModel := e.config.String("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "gpt-oss:20b"
	}
	if c.String("model") != "" {
		ollamaModel = c.String("model")
	}

	slog.Info("Clearing embeddings for model", "model", ollamaModel)

	modelText := pgtype.Text{String: ollamaModel, Valid: true}

	if err := queries.DeleteSpellEmbeddings(c.Context, modelText); err != nil {
		return fmt.Errorf("failed to clear spell embeddings: %w", err)
	}

	if err := queries.DeleteCreatureEmbeddings(c.Context, modelText); err != nil {
		return fmt.Errorf("failed to clear creature embeddings: %w", err)
	}

	if err := queries.DeleteClassEmbeddings(c.Context, modelText); err != nil {
		return fmt.Errorf("failed to clear class embeddings: %w", err)
	}

	if err := queries.DeleteSpeciesEmbeddings(c.Context, modelText); err != nil {
		return fmt.Errorf("failed to clear species embeddings: %w", err)
	}

	slog.Info("Successfully cleared embeddings for model", "model", ollamaModel)
	return nil
}
