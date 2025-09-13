// Package main provides the embeddings CLI application
package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mikeblum/scry.quest/conf"
	"github.com/mikeblum/scry.quest/embeddings"
	"github.com/mikeblum/scry.quest/log"
	"github.com/urfave/cli/v2"
)

// NewEmbeddingsCLI creates a new CLI application for embeddings processing
func NewEmbeddingsCLI() *cli.App {
	return &cli.App{
		Name:  "embeddings",
		Usage: "SRD embeddings pipeline",
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
				Name:   "migrate",
				Usage:  "Run database migrations",
				Action: runMigrations,
			},
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
					return run(c, generateEmbeddings)
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
					return run(c, searchContent)
				},
			},
			{
				Name:  "stats",
				Usage: "Show embedding statistics",
				Action: func(c *cli.Context) error {
					return run(c, showStats)
				},
			},
			{
				Name:  "clear",
				Usage: "Clear embeddings for a model",
				Action: func(c *cli.Context) error {
					return run(c, clearEmbeddings)
				},
			},
		},
	}
}

func run(c *cli.Context, handler func(*cli.Context, *embeddings.Engine) error) error {
	configPath := c.String("config")
	config, err := conf.New(c.Context, &configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	log.NewFromCLI(c)

	engine, cleanup, err := setupServices(c.Context, c, config)
	if err != nil {
		return fmt.Errorf("failed to setup services: %w", err)
	}
	defer func() {
		if cleanup != nil {
			if err := cleanup(); err != nil {
				slog.Error("Failed to cleanup", "error", err)
			}
		}
	}()

	return handler(c, engine)
}

func setupServices(ctx context.Context, c *cli.Context, config *conf.Config) (*embeddings.Engine, func() error, error) {
	queries, conn, cleanup, err := embeddings.NewDBConn(ctx)
	if err != nil {
		return nil, nil, err
	}

	client, err := embeddings.NewOllamaClient(ctx, c, config, cleanup)
	if err != nil {
		return nil, nil, err
	}

	engine := &embeddings.Engine{
		Config:  config,
		Client:  client,
		Queries: queries,
		Conn:    conn,
	}

	return engine, cleanup, nil
}
