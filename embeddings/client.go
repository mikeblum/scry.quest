package embeddings

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ollama/ollama/api"
)

// Client provides Ollama embedding generation.
type Client struct {
	client *api.Client
	config *Config
}

// Config holds Ollama connection settings.
type Config struct {
	Host  string
	Model Model
}

// NewClient creates an Ollama client.
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

// GenerateEmbedding creates embeddings from text.
func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	req := &api.EmbeddingRequest{
		Model:  string(c.config.Model),
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

// Ping tests Ollama server connectivity.
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Heartbeat(ctx)
}
