package embeddings

// Model represents different Ollama models available for embeddings
type Model string

const (
	// Chat model for general purpose text generation
	Chat Model = "gpt-oss:20b"
	// Embedding model for text embeddings
	Embedding Model = "nomic-embed-text"
)

// ModelDimension maps model to dimensions
func ModelDimension(model Model) int {
	switch model {
	case Chat:
		return 1536 // gpt-oss models use 1536 dimensions similar to OpenAI
	case Embedding:
		return 768
	default:
		// Default to gpt-oss:20b dimensions
		return 1536
	}
}
