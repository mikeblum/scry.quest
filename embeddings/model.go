package embeddings

// Model represents Ollama embedding models.
type Model string

const (
	// Chat model used for chat completions (2880-dimensional embeddings)
	Chat Model = "gpt-oss:20b" // 2880-dimensional
	// Embedding model used for text embeddings (768-dimensional embeddings)
	Embedding Model = "nomic-embed-text" // 768-dimensional
)

// ModelDimension returns embedding dimensions.
func ModelDimension(model Model) int {
	switch model {
	case Chat:
		return 2880 // gpt-oss:20b actual embedding length
	case Embedding:
		return 768 // nomic-embed-text embedding length
	default:
		// Default to gpt-oss:20b dimensions
		return 2880
	}
}
