package embeddings

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJSONNameLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "basic creature name",
			line:     `  "name": "Harpy",`,
			expected: "Harpy",
		},
		{
			name:     "creature name with spaces",
			line:     `  "name": "Young Green Dragon",`,
			expected: "Young Green Dragon",
		},
		{
			name:     "creature name without comma",
			line:     `  "name": "Goblin"`,
			expected: "Goblin",
		},
		{
			name:     "no name field",
			line:     `  "type": "humanoid",`,
			expected: "",
		},
		{
			name:     "empty name",
			line:     `  "name": "",`,
			expected: "",
		},
		{
			name:     "name with special characters",
			line:     `  "name": "Orc War Chief",`,
			expected: "Orc War Chief",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseJSONNameLine(tt.line)
			if result != tt.expected {
				t.Errorf("parseJSONNameLine(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestExtractFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		result   *EmbeddingResult
		expected string
	}{
		{
			name: "simple filename",
			result: &EmbeddingResult{
				Metadata: map[string]any{
					"filename": "goblin.json",
				},
			},
			expected: "Goblin",
		},
		{
			name: "filename with underscores",
			result: &EmbeddingResult{
				Metadata: map[string]any{
					"filename": "young_green_dragon.json",
				},
			},
			expected: "Young Green Dragon",
		},
		{
			name: "filename with path",
			result: &EmbeddingResult{
				Metadata: map[string]any{
					"filename": "/path/to/orc_war_chief.json",
				},
			},
			expected: "Orc War Chief",
		},
		{
			name: "no filename metadata",
			result: &EmbeddingResult{
				Metadata: map[string]any{},
			},
			expected: "",
		},
		{
			name: "non-string filename",
			result: &EmbeddingResult{
				Metadata: map[string]any{
					"filename": 123,
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFromFilename(tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFromJSON(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "JSON with name on first line",
			content: `{
  "name": "Harpy",
  "type": "monstrosity"
}`,
			expected: "Harpy",
		},
		{
			name: "JSON with name on second line",
			content: `{
  "index": "harpy",
  "name": "Harpy",
  "type": "monstrosity"
}`,
			expected: "Harpy",
		},
		{
			name: "JSON with name beyond search limit",
			content: `{
  "index": "harpy",
  "type": "monstrosity",
  "alignment": "chaotic evil",
  "armor_class": 11,
  "hit_points": 38,
  "name": "Harpy"
}`,
			expected: "",
		},
		{
			name:     "not JSON content",
			content:  "This is not JSON",
			expected: "",
		},
		{
			name: "JSON without name field",
			content: `{
  "type": "humanoid",
  "alignment": "chaotic evil"
}`,
			expected: "",
		},
		{
			name: "JSON with empty name",
			content: `{
  "name": "",
  "type": "humanoid"
}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFromJSON(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFromMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "markdown with single hash",
			content:  "# Fireball\nA bright streak flashes from your pointing finger...",
			expected: "Fireball",
		},
		{
			name:     "markdown with multiple hashes",
			content:  "## Lightning Bolt\nA stroke of lightning forming a line...",
			expected: "Lightning Bolt",
		},
		{
			name:     "markdown with spaces after hashes",
			content:  "#    Magic Missile   \nYou create three glowing darts...",
			expected: "Magic Missile",
		},
		{
			name:     "not markdown content",
			content:  "This is not markdown",
			expected: "",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "markdown with empty header",
			content:  "#\nSome content here",
			expected: "",
		},
		{
			name:     "markdown starting with other content",
			content:  "Some text\n# Header\nMore content",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFromMarkdown(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnsureValidJSON(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected string
		wantErr  bool
	}{
		{
			name:     "already valid JSON object",
			content:  []byte(`{"name": "Goblin", "type": "humanoid"}`),
			expected: `{"name": "Goblin", "type": "humanoid"}`,
			wantErr:  false,
		},
		{
			name:     "already valid JSON array",
			content:  []byte(`["spell1", "spell2"]`),
			expected: `["spell1", "spell2"]`,
			wantErr:  false,
		},
		{
			name:     "plain text content",
			content:  []byte(`This is plain text content`),
			expected: `{"content":"This is plain text content"}`,
			wantErr:  false,
		},
		{
			name:     "markdown content",
			content:  []byte("# Title\nSome markdown content"),
			expected: `{"content":"# Title\nSome markdown content"}`,
			wantErr:  false,
		},
		{
			name:     "empty content",
			content:  []byte(``),
			expected: `{"content":""}`,
			wantErr:  false,
		},
		{
			name:     "malformed JSON",
			content:  []byte(`{"name": "Goblin", "type": }`),
			expected: `{"content":"{\"name\": \"Goblin\", \"type\": }"}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ensureValidJSON(tt.content)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.JSONEq(t, tt.expected, string(result))
			}
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name         string
		metadata     map[string]any
		key          string
		defaultValue string
		expected     string
	}{
		{
			name: "key exists with string value",
			metadata: map[string]any{
				"model": "nomic-embed-text",
				"count": 42,
			},
			key:          "model",
			defaultValue: "default-model",
			expected:     "nomic-embed-text",
		},
		{
			name: "key exists with non-string value",
			metadata: map[string]any{
				"model": 123,
				"count": 42,
			},
			key:          "model",
			defaultValue: "default-model",
			expected:     "default-model",
		},
		{
			name: "key does not exist",
			metadata: map[string]any{
				"count": 42,
			},
			key:          "model",
			defaultValue: "default-model",
			expected:     "default-model",
		},
		{
			name:         "empty metadata",
			metadata:     map[string]any{},
			key:          "model",
			defaultValue: "default-model",
			expected:     "default-model",
		},
		{
			name:         "nil metadata",
			metadata:     nil,
			key:          "model",
			defaultValue: "default-model",
			expected:     "default-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.metadata, tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractSimpleName(t *testing.T) {
	tests := []struct {
		name     string
		result   *EmbeddingResult
		content  string
		expected string
	}{
		{
			name: "extract from filename",
			result: &EmbeddingResult{
				ContentID: uuid.New(),
				Metadata: map[string]any{
					"filename": "fire_elemental.json",
				},
			},
			content:  `{"type": "elemental"}`,
			expected: "Fire Elemental",
		},
		{
			name: "extract from JSON when filename fails",
			result: &EmbeddingResult{
				ContentID: uuid.New(),
				Metadata:  map[string]any{},
			},
			content: `{
  "name": "Ancient Red Dragon",
  "type": "dragon"
}`,
			expected: "Ancient Red Dragon",
		},
		{
			name: "extract from markdown when filename and JSON fail",
			result: &EmbeddingResult{
				ContentID: uuid.New(),
				Metadata:  map[string]any{},
			},
			content:  "# Wish\nThe mightiest spell a mortal creature can cast...",
			expected: "Wish",
		},
		{
			name: "fallback to empty string when all extraction methods fail",
			result: &EmbeddingResult{
				ContentID: uuid.New(),
				Metadata:  map[string]any{},
			},
			content:  "This is plain text with no extractable name",
			expected: "",
		},
		{
			name: "filename takes precedence over JSON name",
			result: &EmbeddingResult{
				ContentID: uuid.New(),
				Metadata: map[string]any{
					"filename": "goblin_shaman.json",
				},
			},
			content: `{
  "name": "Different Name",
  "type": "humanoid"
}`,
			expected: "Goblin Shaman",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSimpleName(tt.result, tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseBestiaryData(t *testing.T) {
	tests := []struct {
		name              string
		content           []byte
		expectedSize      string
		expectedType      string
		expectedAlignment string
		expectedLanguages string
		expectedSenses    string
	}{
		{
			name: "goblin warrior data",
			content: []byte(`{
  "name": "Goblin Warrior",
  "size": "Small",
  "type": "Fey (Goblinoid)",
  "alignment": "Chaotic Neutral",
  "ability_scores": {
    "strength": { "score": 8, "modifier": -1 }
  },
  "skills": [
    { "name": "Stealth", "modifier": 6 }
  ],
  "speed": {
    "walk": { "value": 30, "unit": "ft." }
  },
  "languages": ["Common", "Goblin"],
  "senses": {
    "darkvision": { "value": 60, "unit": "ft." },
    "passive_perception": 9
  }
}`),
			expectedSize:      "Small",
			expectedType:      "Fey (Goblinoid)",
			expectedAlignment: "Chaotic Neutral",
			expectedLanguages: "Common, Goblin",
			expectedSenses:    "passive Perception 9, darkvision 60 ft.",
		},
		{
			name: "invalid JSON defaults",
			content: []byte(`{invalid json`),
			expectedSize:      "Medium",
			expectedType:      "",
			expectedAlignment: "Unaligned",
			expectedLanguages: "—",
			expectedSenses:    "—",
		},
		{
			name: "minimal creature data",
			content: []byte(`{
  "name": "Simple Creature",
  "size": "Large"
}`),
			expectedSize:      "Large",
			expectedType:      "",
			expectedAlignment: "Unaligned",
			expectedLanguages: "—",
			expectedSenses:    "—",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, creatureType, alignment, abilities, skills, speed, languages, senses := parseBestiaryData(tt.content)

			assert.Equal(t, tt.expectedSize, size)
			assert.Equal(t, tt.expectedType, creatureType)
			assert.Equal(t, tt.expectedAlignment, alignment)
			assert.Equal(t, tt.expectedLanguages, languages)

			// For goblin warrior, check specific senses (order may vary due to map iteration)
			if tt.name == "goblin warrior data" {
				assert.Contains(t, senses, "passive Perception 9")
				assert.Contains(t, senses, "darkvision 60 ft.")
			} else {
				assert.Equal(t, tt.expectedSenses, senses)
			}

			// Verify JSON fields are valid JSON
			assert.True(t, json.Valid(abilities), "abilities should be valid JSON")
			assert.True(t, json.Valid(skills), "skills should be valid JSON")
			assert.True(t, json.Valid(speed), "speed should be valid JSON")
		})
	}
}

func TestParseBestiaryDataWithActualFiles(t *testing.T) {
	testFiles := []struct {
		filename          string
		expectedSize      string
		expectedType      string
		expectedAlignment string
	}{
		{
			filename:          "../srd/beastiary/goblin_warrior.json",
			expectedSize:      "Small",
			expectedType:      "Fey (Goblinoid)",
			expectedAlignment: "Chaotic Neutral",
		},
		{
			filename:          "../srd/beastiary/dragon_black_adult.json",
			expectedSize:      "Huge",
			expectedType:      "Dragon (Chromatic)",
			expectedAlignment: "Chaotic Evil",
		},
	}

	for _, tt := range testFiles {
		t.Run(tt.filename, func(t *testing.T) {
			// Read the actual file
			content, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Skipf("Could not read test file %s: %v", tt.filename, err)
				return
			}

			size, creatureType, alignment, abilities, skills, speed, languages, senses := parseBestiaryData(content)

			assert.Equal(t, tt.expectedSize, size)
			assert.Equal(t, tt.expectedType, creatureType)
			assert.Equal(t, tt.expectedAlignment, alignment)

			// Verify JSON fields are valid JSON and not empty
			assert.True(t, json.Valid(abilities), "abilities should be valid JSON")
			assert.True(t, json.Valid(skills), "skills should be valid JSON")
			assert.True(t, json.Valid(speed), "speed should be valid JSON")

			// Verify we got non-default values
			assert.NotEqual(t, "—", languages, "should extract actual languages")
			assert.NotEqual(t, "—", senses, "should extract actual senses")

			t.Logf("Parsed creature: size=%s, type=%s, alignment=%s", size, creatureType, alignment)
			t.Logf("Languages: %s", languages)
			t.Logf("Senses: %s", senses)
		})
	}
}

func TestParseSpellData(t *testing.T) {
	tests := []struct {
		name                 string
		content              []byte
		expectedLevel        int
		expectedSchool       string
		expectedCastingTime  string
		expectedRange        string
		expectedComponents   string
		expectedDuration     string
		expectedClasses      []string
	}{
		{
			name: "fireball spell data",
			content: []byte(`{
  "name": "Fireball",
  "level": "Level 3",
  "school": "Evocation",
  "casting_time": {
    "value": "1",
    "unit": "Action"
  },
  "range": {
    "value": 150,
    "unit": "feet"
  },
  "components": {
    "verbal": true,
    "somatic": true,
    "material": "a tiny ball of bat guano and sulfur"
  },
  "duration": {
    "type": "instantaneous"
  },
  "classes": ["Sorcerer", "Wizard"]
}`),
			expectedLevel:       3,
			expectedSchool:      "Evocation",
			expectedCastingTime: "1 Action",
			expectedRange:       "150 feet",
			expectedComponents:  "V, S, M (a tiny ball of bat guano and sulfur)",
			expectedDuration:    "Instantaneous",
			expectedClasses:     []string{"Sorcerer", "Wizard"},
		},
		{
			name: "cantrip spell data",
			content: []byte(`{
  "name": "Sacred Flame",
  "level": "Cantrip",
  "school": "Evocation",
  "casting_time": {
    "value": "1",
    "unit": "Action"
  },
  "range": {
    "value": 60,
    "unit": "feet"
  },
  "components": {
    "verbal": true,
    "somatic": true
  },
  "duration": {
    "type": "instantaneous"
  }
}`),
			expectedLevel:       0,
			expectedSchool:      "Evocation",
			expectedCastingTime: "1 Action",
			expectedRange:       "60 feet",
			expectedComponents:  "V, S",
			expectedDuration:    "Instantaneous",
			expectedClasses:     []string{},
		},
		{
			name: "invalid JSON defaults",
			content: []byte(`{invalid json`),
			expectedLevel:       0,
			expectedSchool:      "N/A",
			expectedCastingTime: "N/A",
			expectedRange:       "N/A",
			expectedComponents:  "N/A",
			expectedDuration:    "N/A",
			expectedClasses:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, school, castingTime, rangeValue, components, duration, classes := parseSpellData(tt.content)

			assert.Equal(t, tt.expectedLevel, level)
			assert.Equal(t, tt.expectedSchool, school)
			assert.Equal(t, tt.expectedCastingTime, castingTime)
			assert.Equal(t, tt.expectedRange, rangeValue)
			assert.Equal(t, tt.expectedComponents, components)
			assert.Equal(t, tt.expectedDuration, duration)
			assert.Equal(t, tt.expectedClasses, classes)
		})
	}
}

func TestParseSpellDataWithActualFiles(t *testing.T) {
	testFiles := []struct {
		filename            string
		expectedLevel       int
		expectedSchool      string
		expectNonDefaults   bool
	}{
		{
			filename:          "../srd/spells/fireball.json",
			expectedLevel:     3,
			expectedSchool:    "Evocation",
			expectNonDefaults: true,
		},
		{
			filename:          "../srd/spells/sacred_flame.json",
			expectedLevel:     0,
			expectedSchool:    "Evocation",
			expectNonDefaults: true,
		},
	}

	for _, tt := range testFiles {
		t.Run(tt.filename, func(t *testing.T) {
			// Read the actual file
			content, err := os.ReadFile(tt.filename)
			if err != nil {
				t.Skipf("Could not read test file %s: %v", tt.filename, err)
				return
			}

			level, school, castingTime, rangeValue, components, duration, classes := parseSpellData(content)

			assert.Equal(t, tt.expectedLevel, level)
			assert.Equal(t, tt.expectedSchool, school)

			if tt.expectNonDefaults {
				// Verify we got non-default values
				assert.NotEqual(t, "N/A", castingTime, "should extract actual casting time")
				assert.NotEqual(t, "N/A", rangeValue, "should extract actual range")
				assert.NotEqual(t, "N/A", components, "should extract actual components")
				assert.NotEqual(t, "N/A", duration, "should extract actual duration")
			}

			t.Logf("Parsed spell: level=%d, school=%s", level, school)
			t.Logf("Casting time: %s, Range: %s", castingTime, rangeValue)
			t.Logf("Components: %s, Duration: %s", components, duration)
			t.Logf("Classes: %v", classes)
		})
	}
}