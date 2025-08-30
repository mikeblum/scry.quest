package embeddings //nolint:revive // package comment not needed

// ContentType represents different types of D&D content
type ContentType string

const (
	// ContentTypeSpell represents spell content
	ContentTypeSpell ContentType = "spell"
	// ContentTypeBestiary represents bestiary content
	ContentTypeBestiary ContentType = "bestiary"
	// ContentTypeClass represents class content
	ContentTypeClass ContentType = "class"
	// ContentTypeSpecies represents species content
	ContentTypeSpecies ContentType = "species"
)

// SearchResult represents a search result with similarity score
type SearchResult struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Type       ContentType `json:"type"`
	Content    string      `json:"content"`
	Similarity float64     `json:"similarity"`
}

// SearchOptions configures search parameters
type SearchOptions struct {
	ContentTypes []ContentType `json:"content_types"`
	Limit        int32         `json:"limit"`
	Threshold    float64       `json:"threshold"`
}
