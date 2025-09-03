// Package embeddings generates, stores, and searches vector embeddings for SRD content.
// Uses Ollama for embedding generation and PostgreSQL with pgvector for storage.
//
//	client, _ := NewClient(Config{Host: "http://localhost:11434", Model: Embedding})
//	pipeline := NewPipeline(client, processor, store)
//	ProcessAllSRDContent(ctx, pipeline, "./srd")
//	results, _ := NewSearchService(client, queries).Search(ctx, "fire spell", nil)
package embeddings
