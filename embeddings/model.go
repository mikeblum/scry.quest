package embeddings

// Model represents Ollama embedding models.
type Model string

const (
	// Chat model used for chat completions (2880-dimensional embeddings)
	Chat Model = "gpt-oss:20b"
	// Embedding model used for text embeddings (768-dimensional embeddings)
	Embedding Model = "nomic-embed-text"
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

// ModelContextLength returns maximum context length in tokens.
func ModelContextLength(model Model) int {
	switch model {
	case Chat:
		return 4096 // gpt-oss:20b context length
	case Embedding:
		return 8192 // nomic-embed-text context length
	default:
		// Default to gpt-oss:20b context length
		return 4096
	}
}
