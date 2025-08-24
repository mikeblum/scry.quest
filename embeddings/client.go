package embeddings

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/ollama/ollama/api"
)

type Client struct {
	client *api.Client
	config *Config
}

type Config struct {
	Host  string `env:"OLLAMA_HOST" env-default:"http://localhost:11434"`
	Model string `env:"OLLAMA_MODEL" env-default:"nomic-embed-text"`
}

func NewClient(cfg Config) (*Client, error) {
	_, err := url.Parse(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid Ollama host URL: %w", err)
	}

	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	return &Client{
		client: client,
		config: &cfg,
	}, nil
}

// GenerateEmbedding creates an embedding vector for the given text
func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	req := &api.EmbeddingRequest{
		Model:  c.model,
		Prompt: text,
	}

	resp, err := c.client.Embeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Embedding) == 0 {
		return nil, fmt.Errorf("received empty embedding response")
	}

	// Convert from []float64 to []float32
	result := make([]float32, len(resp.Embedding))
	for i, v := range resp.Embedding {
		result[i] = float32(v)
	}

	return result, nil
}

// GetModelDimensions returns the expected dimensions for the current model
func (c *Client) GetModelDimensions() int {
	switch c.model {
	case "gpt-oss:20b":
		return 1536 // gpt-oss models use 1536 dimensions similar to OpenAI
	case "nomic-embed-text":
		return 768
	default:
		// Default to gpt-oss:20b dimensions
		return 1536
	}
}
