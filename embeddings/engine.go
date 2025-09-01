package embeddings

import (
	"github.com/mikeblum/scry.quest/conf"
	"github.com/mikeblum/scry.quest/internal/database"
)

// Engine represents the embeddings processing engine
type Engine struct {
	Config  *conf.Config
	Client  *Client
	Queries *database.Queries
}

// NewEngine creates a new embeddings processing engine
func NewEngine() *Engine {
	return &Engine{}
}
