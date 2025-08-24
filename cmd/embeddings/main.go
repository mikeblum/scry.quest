package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mikeblum/scry.quest/conf"
	"github.com/mikeblum/scry.quest/embeddings"
	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/mikeblum/scry.quest/log"
)

func main() {
	var (
		configPath  = flag.String("config", "env/.env.local", "Path to configuration file")
		command     = flag.String("command", "", "Command to run: generate, search, stats, clear")
		model       = flag.String("model", "gpt-oss:20b", "Embedding model to use (gpt-oss:20b, nomic-embed-text, all-minilm)")
		ollamaURL   = flag.String("ollama-url", "http://localhost:11434", "Ollama server URL")
		contentType = flag.String("type", "", "Content type filter (spell, bestiary, class, species)")
		query       = flag.String("query", "", "Search query (for search command)")
		limit       = flag.Int("limit", 10, "Number of results to return (for search command)")
		srdPath     = flag.String("srd", "./srd", "Path to SRD directory")
	)
	flag.Parse()

	if *command == "" {
		fmt.Println("Usage: embeddings -command <generate|search|stats|clear> [options]")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  generate    Generate embeddings for SRD content")
		fmt.Println("  search      Search content using embeddings")
		fmt.Println("  stats       Show embedding statistics")
		fmt.Println("  clear       Clear embeddings for a model")
		fmt.Println("")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Initialize logging
	log.NewFromEnv()

	ctx := context.Background()

	// Load configuration
	cfg, err := conf.New(ctx, *configPath)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize database connection
	conn, err := pgx.Connect(ctx, cfg.String("DATABASE_URL"))
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	queries := database.New(conn)

	// Initialize Ollama client
	ollamaConfig := embeddings.Config{
		Host:  *ollamaURL,
		Model: *model,
	}
	client, err := embeddings.NewClient(ollamaConfig)
	if err != nil {
		slog.Error("Failed to create Ollama client", "error", err)
		os.Exit(1)
	}

	// Test Ollama connection
	if err := client.Ping(ctx); err != nil {
		slog.Error("Failed to connect to Ollama server", "url", *ollamaURL, "error", err)
		os.Exit(1)
	}

	fmt.Printf("Connected to Ollama server at %s using model %s\n", *ollamaURL, *model)

	switch *command {
	case "generate":
		if err := generateEmbeddings(ctx, client, queries, *srdPath, *contentType); err != nil {
			slog.Error("Failed to generate embeddings", "error", err)
			os.Exit(1)
		}

	case "search":
		if *query == "" {
			slog.Error("Query is required for search command")
			os.Exit(1)
		}
		if err := searchContent(ctx, client, queries, *query, *contentType, *limit); err != nil {
			slog.Error("Failed to search content", "error", err)
			os.Exit(1)
		}

	case "stats":
		if err := showStats(ctx, queries); err != nil {
			slog.Error("Failed to show stats", "error", err)
			os.Exit(1)
		}

	case "clear":
		if err := clearEmbeddings(ctx, queries, *model); err != nil {
			slog.Error("Failed to clear embeddings", "error", err)
			os.Exit(1)
		}

	default:
		slog.Error("Unknown command", "command", *command)
		os.Exit(1)
	}
}

func generateEmbeddings(ctx context.Context, client *embeddings.Client, queries *database.Queries, srdPath, contentType string) error {
	pipeline := embeddings.NewPipeline(client, queries, srdPath)

	if contentType == "" {
		fmt.Println("Generating embeddings for all content types...")
		return pipeline.ProcessAll(ctx)
	}

	// Process specific content type
	switch contentType {
	case "spell":
		return pipeline.ProcessContentType(ctx, embeddings.ContentTypeSpell)
	case "bestiary":
		return pipeline.ProcessContentType(ctx, embeddings.ContentTypeBestiary)
	case "class":
		return pipeline.ProcessContentType(ctx, embeddings.ContentTypeClass)
	case "species":
		return pipeline.ProcessContentType(ctx, embeddings.ContentTypeSpecies)
	default:
		return fmt.Errorf("unknown content type: %s", contentType)
	}
}

func searchContent(ctx context.Context, client *embeddings.Client, queries *database.Queries, query, contentType string, limit int) error {
	searchService := embeddings.NewSearchService(client, queries)

	var results []*embeddings.SearchResult
	var err error

	if contentType == "" {
		// Search all content types
		results, err = searchService.Search(ctx, query, &embeddings.SearchOptions{
			ContentTypes: []embeddings.ContentType{
				embeddings.ContentTypeSpell,
				embeddings.ContentTypeBestiary,
				embeddings.ContentTypeClass,
				embeddings.ContentTypeSpecies,
			},
			Limit:     int32(limit),
			Threshold: 0.6,
		})
	} else {
		// Search specific content type
		switch contentType {
		case "spell":
			results, err = searchService.SearchSpells(ctx, query, int32(limit))
		case "bestiary":
			results, err = searchService.SearchBestiary(ctx, query, int32(limit))
		case "class":
			results, err = searchService.SearchClasses(ctx, query, int32(limit))
		case "species":
			results, err = searchService.SearchSpecies(ctx, query, int32(limit))
		default:
			return fmt.Errorf("unknown content type: %s", contentType)
		}
	}

	if err != nil {
		return err
	}

	fmt.Printf("Found %d results for query: %s\n\n", len(results), query)

	for i, result := range results {
		fmt.Printf("%d. %s (%s) - Similarity: %.3f\n", i+1, result.Name, result.Type, result.Similarity)
		if result.Content != "" && len(result.Content) > 100 {
			fmt.Printf("   %s...\n", result.Content[:100])
		} else if result.Content != "" {
			fmt.Printf("   %s\n", result.Content)
		}
		fmt.Println()
	}

	return nil
}

func showStats(ctx context.Context, queries *database.Queries) error {
	stats, err := queries.GetEmbeddingStats(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Embedding Statistics:")
	fmt.Println("=====================")
	fmt.Printf("%-12s %-20s %-12s %-12s %-12s\n", "Table", "Model", "Total", "Embedded", "Expected Dims")
	fmt.Println(strings.Repeat("-", 80))

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

		fmt.Printf("%-12s %-20s %-12d %-12d %-12s\n",
			stat.TableName,
			embeddingModel,
			stat.TotalRows,
			stat.EmbeddedRows,
			expectedDims)
	}

	// Show counts by model
	fmt.Println("\nCounts by Model:")
	fmt.Println("================")

	counts, err := queries.CountItemsByEmbeddingModel(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%-12s %-20s %-8s\n", "Table", "Model", "Count")
	fmt.Println(strings.Repeat("-", 45))

	for _, count := range counts {
		fmt.Printf("%-12s %-20s %-8d\n", count.TableName, count.Model, count.Count)
	}

	return nil
}

func clearEmbeddings(ctx context.Context, queries *database.Queries, model string) error {
	fmt.Printf("Clearing embeddings for model: %s\n", model)

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

	fmt.Printf("Successfully cleared embeddings for model: %s\n", model)
	return nil
}
