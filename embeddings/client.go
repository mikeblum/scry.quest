package embeddings //nolint:revive // package comment not needed

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ollama/ollama/api"
)

const (
	defaultOllamaModel = "gpt-oss:20b"
)

// Client provides Ollama API functionality for generating embeddings
type Client struct {
	client *api.Client
	config *Config
}

// Config holds configuration for the Ollama client
type Config struct {
	Host  string
	Model string
}

// NewClient creates a new Ollama client with the given configuration
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
		Model:  c.config.Model,
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

// Ping tests the connection to the Ollama server
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping Ollama server: %w", err)
	}
	return nil
}

// GetModelDimensions returns the expected dimensions for the current model
func (c *Client) GetModelDimensions() int {
	switch c.config.Model {
	case defaultOllamaModel:
		return 1536 // gpt-oss models use 1536 dimensions similar to OpenAI
	case "nomic-embed-text":
		return 768
	default:
		// Default to gpt-oss:20b dimensions
		return 1536
	}
}
