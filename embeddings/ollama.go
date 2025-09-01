package embeddings

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mikeblum/scry.quest/conf"
	"github.com/urfave/cli/v2"
)

const (
	// EnvOllamaHost is the environment variable for Ollama server host
	EnvOllamaHost = "OLLAMA_HOST"
	// EnvOllamaModel is the environment variable for Ollama model name
	EnvOllamaModel = "OLLAMA_MODEL"
	// FlagOllamaHost is the CLI flag name for Ollama server host
	FlagOllamaHost = "ollama-host"
	// FlagOllamaModel is the CLI flag name for Ollama model
	FlagOllamaModel   = "ollama-model"
	ollamaDefaultHost = "http://localhost:11434"
)

// NewOllamaClient creates a new Ollama client with configuration
func NewOllamaClient(ctx context.Context, c *cli.Context, _ *conf.Config, cleanup func() error) (*Client, error) {
	conf, err := conf.New(ctx, nil)
	if err != nil {
		return nil, err
	}
	ollamaHost := conf.FromCLI(c, FlagOllamaHost, EnvOllamaHost, ollamaDefaultHost)
	ollamaModel := conf.FromCLI(c, FlagOllamaModel, EnvOllamaModel, string(Embedding))

	client, err := NewClient(Config{
		Host:  ollamaHost,
		Model: Model(ollamaModel),
	})
	if err != nil {
		if cleanupErr := cleanup(); cleanupErr != nil {
			slog.Error("Failed to cleanup during client creation", "error", cleanupErr)
		}
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	// Test Ollama connection
	if err := client.Ping(ctx); err != nil {
		if cleanupErr := cleanup(); cleanupErr != nil {
			slog.Error("Failed to cleanup during ping failure", "error", cleanupErr)
		}
		return nil, fmt.Errorf("failed to connect to Ollama server at %s: %w", ollamaHost, err)
	}

	slog.Info("Connected to Ollama server", "url", ollamaHost, "model", ollamaModel)
	return client, nil
}
