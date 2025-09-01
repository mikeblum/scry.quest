package embeddings

// ContentType represents SRD content categories.
type ContentType string

const (
	// ContentTypeSpell ex. Fireball
	ContentTypeSpell ContentType = "spell"
	// ContentTypeBestiary ex. Kobolds
	ContentTypeBestiary ContentType = "bestiary"
	// ContentTypeClass ex. Monk
	ContentTypeClass ContentType = "class"
	// ContentTypeSpecies ex. Dwarf
	ContentTypeSpecies ContentType = "species"
)
