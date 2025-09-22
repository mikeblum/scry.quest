package embeddings

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mikeblum/scry.quest/internal/database"
	"github.com/pgvector/pgvector-go"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// DatabaseEmbeddingStore stores embeddings in PostgreSQL.
type DatabaseEmbeddingStore struct {
	queries *database.Queries
	conn    *pgx.Conn
}

// NewDatabaseEmbeddingStore creates a database store with connection for transactions.
func NewDatabaseEmbeddingStore(queries *database.Queries, conn *pgx.Conn) *DatabaseEmbeddingStore {
	return &DatabaseEmbeddingStore{
		queries: queries,
		conn:    conn,
	}
}

// Store implements the EmbeddingStore interface
func (s *DatabaseEmbeddingStore) Store(ctx context.Context, result *EmbeddingResult) error {
	vector := pgvector.NewVector(result.Embedding)
	model := getString(result.Metadata, "model", "nomic-embed-text")
	modelText := pgtype.Text{String: model, Valid: true}

	originalContent := result.Metadata["original_content"].([]byte)
	description := string(originalContent)
	if len(description) > 1000 {
		description = description[:1000]
	}

	// Convert content to valid JSON for storage
	rawData, err := ensureValidJSON(originalContent)
	if err != nil {
		return fmt.Errorf("failed to convert content to JSON: %w", err)
	}

	// Extract a simple name from the content or use filename
	name := extractSimpleName(result, string(originalContent))
	if name == "" {
		name = result.ContentID.String() // Use UUID as fallback
	}

	switch result.ContentType {
	case string(ContentTypeSpell):
		params := createSpellParamsFromData(name, description, vector, rawData, modelText, originalContent)
		_, err := s.queries.CreateSpell(ctx, params)
		return err

	case string(ContentTypeBestiary):
		params := createCreatureParamsFromData(name, vector, rawData, modelText, originalContent)
		_, err := s.queries.CreateCreature(ctx, params)
		return err

	case string(ContentTypeClass):
		params := database.CreateClassParams{
			Name:           name,
			Description:    pgtype.Text{String: description, Valid: true},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		}
		_, err := s.queries.CreateClass(ctx, params)
		return err

	case string(ContentTypeSpecies):
		params := database.CreateSpeciesParams{
			Name:           name,
			Description:    pgtype.Text{String: description, Valid: true},
			Traits:         []string{},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		}
		_, err := s.queries.CreateSpecies(ctx, params)
		return err

	default:
		return fmt.Errorf("unsupported content type: %s", result.ContentType)
	}
}

// StoreAll implements the EmbeddingStore interface with batch transaction support
func (s *DatabaseEmbeddingStore) StoreAll(ctx context.Context, results []*EmbeddingResult) error {
	return s.batchInsertTx(ctx, results)
}

// batchInsertTx performs efficient batch insertion using sqlc within a transaction
func (s *DatabaseEmbeddingStore) batchInsertTx(ctx context.Context, results []*EmbeddingResult) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txQueries := s.queries.WithTx(tx)

	if err := s.batchInsertByType(ctx, txQueries, results); err != nil {
		return fmt.Errorf("failed to batch insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// batchInsertByType groups results by content type and performs batch inserts
func (s *DatabaseEmbeddingStore) batchInsertByType(ctx context.Context, queries *database.Queries, results []*EmbeddingResult) error {
	grouped := s.groupResultsByType(results)

	if err := s.insertSpells(ctx, queries, grouped.spells); err != nil {
		return err
	}
	if err := s.insertCreatures(ctx, queries, grouped.creatures); err != nil {
		return err
	}
	if err := s.insertClasses(ctx, queries, grouped.classes); err != nil {
		return err
	}
	if err := s.insertSpecies(ctx, queries, grouped.species); err != nil {
		return err
	}

	return nil
}

type groupedResults struct {
	spells    []database.CreateSpellsParams
	creatures []database.CreateCreaturesParams
	classes   []database.CreateClassesParams
	species   []database.CreateSpeciesBatchParams
}

func (s *DatabaseEmbeddingStore) groupResultsByType(results []*EmbeddingResult) groupedResults {
	var grouped groupedResults

	for _, result := range results {
		params := s.createParamsFromResult(result)

		switch result.ContentType {
		case string(ContentTypeSpell):
			grouped.spells = append(grouped.spells, params.spell)
		case string(ContentTypeBestiary):
			grouped.creatures = append(grouped.creatures, params.creature)
		case string(ContentTypeClass):
			grouped.classes = append(grouped.classes, params.class)
		case string(ContentTypeSpecies):
			grouped.species = append(grouped.species, params.species)
		}
	}

	return grouped
}

type batchParams struct {
	spell    database.CreateSpellsParams
	creature database.CreateCreaturesParams
	class    database.CreateClassesParams
	species  database.CreateSpeciesBatchParams
}

func (s *DatabaseEmbeddingStore) createParamsFromResult(result *EmbeddingResult) batchParams {
	vector := pgvector.NewVector(result.Embedding)
	model := getString(result.Metadata, "model", "nomic-embed-text")
	modelText := pgtype.Text{String: model, Valid: true}

	originalContent := result.Metadata["original_content"].([]byte)
	description := string(originalContent)
	if len(description) > 1000 {
		description = description[:1000]
	}

	rawData, _ := ensureValidJSON(originalContent)
	name := extractSimpleName(result, string(originalContent))
	if name == "" {
		name = result.ContentID.String()
	}

	return batchParams{
		spell: createSpellsBatchParamsFromData(name, description, vector, rawData, modelText, originalContent),
		creature: createCreaturesBatchParamsFromData(name, vector, rawData, modelText, originalContent),
		class: database.CreateClassesParams{
			Name:           name,
			Description:    pgtype.Text{String: description, Valid: true},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		},
		species: database.CreateSpeciesBatchParams{
			Name:           name,
			Description:    pgtype.Text{String: description, Valid: true},
			Traits:         []string{},
			Embedding:      vector,
			RawData:        rawData,
			EmbeddingModel: modelText,
		},
	}
}

func (s *DatabaseEmbeddingStore) insertSpells(ctx context.Context, queries *database.Queries, params []database.CreateSpellsParams) error {
	if len(params) > 0 {
		br := queries.CreateSpells(ctx, params)
		var batchErr error
		br.Exec(func(i int, err error) {
			if err != nil && batchErr == nil {
				batchErr = fmt.Errorf("failed to insert spell at index %d: %w", i, err)
			}
		})
		if batchErr != nil {
			return batchErr
		}
	}
	return nil
}

func (s *DatabaseEmbeddingStore) insertCreatures(ctx context.Context, queries *database.Queries, params []database.CreateCreaturesParams) error {
	if len(params) > 0 {
		br := queries.CreateCreatures(ctx, params)
		var batchErr error
		br.Exec(func(i int, err error) {
			if err != nil && batchErr == nil {
				batchErr = fmt.Errorf("failed to insert creature at index %d: %w", i, err)
			}
		})
		if batchErr != nil {
			return batchErr
		}
	}
	return nil
}

func (s *DatabaseEmbeddingStore) insertClasses(ctx context.Context, queries *database.Queries, params []database.CreateClassesParams) error {
	if len(params) > 0 {
		br := queries.CreateClasses(ctx, params)
		var batchErr error
		br.Exec(func(i int, err error) {
			if err != nil && batchErr == nil {
				batchErr = fmt.Errorf("failed to insert class at index %d: %w", i, err)
			}
		})
		if batchErr != nil {
			return batchErr
		}
	}
	return nil
}

func (s *DatabaseEmbeddingStore) insertSpecies(ctx context.Context, queries *database.Queries, params []database.CreateSpeciesBatchParams) error {
	if len(params) > 0 {
		br := queries.CreateSpeciesBatch(ctx, params)
		var batchErr error
		br.Exec(func(i int, err error) {
			if err != nil && batchErr == nil {
				batchErr = fmt.Errorf("failed to insert species at index %d: %w", i, err)
			}
		})
		if batchErr != nil {
			return batchErr
		}
	}
	return nil
}

// createSpellParamsFromData creates spell parameters with parsed JSON data
func createSpellParamsFromData(name, description string, vector pgvector.Vector, rawData []byte, modelText pgtype.Text, originalContent []byte) database.CreateSpellParams {
	level, school, castingTime, rangeValue, components, duration, classes := parseSpellData(originalContent)

	return database.CreateSpellParams{
		Name:           name,
		Description:    pgtype.Text{String: description, Valid: true},
		Level:          safeIntToInt32(level),
		School:         pgtype.Text{String: school, Valid: true},
		CastingTime:    pgtype.Text{String: castingTime, Valid: true},
		RangeValue:     pgtype.Text{String: rangeValue, Valid: true},
		Components:     pgtype.Text{String: components, Valid: true},
		Duration:       pgtype.Text{String: duration, Valid: true},
		Classes:        classes,
		Embedding:      vector,
		RawData:        rawData,
		EmbeddingModel: modelText,
	}
}

// createSpellsBatchParamsFromData creates batch spell parameters with parsed JSON data
func createSpellsBatchParamsFromData(name, description string, vector pgvector.Vector, rawData []byte, modelText pgtype.Text, originalContent []byte) database.CreateSpellsParams {
	level, school, castingTime, rangeValue, components, duration, classes := parseSpellData(originalContent)

	return database.CreateSpellsParams{
		Name:           name,
		Description:    pgtype.Text{String: description, Valid: true},
		Level:          safeIntToInt32(level),
		School:         pgtype.Text{String: school, Valid: true},
		CastingTime:    pgtype.Text{String: castingTime, Valid: true},
		RangeValue:     pgtype.Text{String: rangeValue, Valid: true},
		Components:     pgtype.Text{String: components, Valid: true},
		Duration:       pgtype.Text{String: duration, Valid: true},
		Classes:        classes,
		Embedding:      vector,
		RawData:        rawData,
		EmbeddingModel: modelText,
	}
}

// createCreatureParamsFromData creates creature parameters with parsed JSON data
func createCreatureParamsFromData(name string, vector pgvector.Vector, rawData []byte, modelText pgtype.Text, originalContent []byte) database.CreateCreatureParams {
	size, creatureType, alignment, abilities, skills, speed, languages, senses := parseBestiaryData(originalContent)

	return database.CreateCreatureParams{
		Name:           name,
		Size:           pgtype.Text{String: size, Valid: true},
		Type:           pgtype.Text{String: creatureType, Valid: creatureType != ""},
		Alignment:      pgtype.Text{String: alignment, Valid: true},
		Abilities:      abilities,
		Skills:         skills,
		Speed:          speed,
		Languages:      pgtype.Text{String: languages, Valid: true},
		Senses:         pgtype.Text{String: senses, Valid: true},
		Embedding:      vector,
		RawData:        rawData,
		EmbeddingModel: modelText,
	}
}

// createCreaturesBatchParamsFromData creates batch creature parameters with parsed JSON data
func createCreaturesBatchParamsFromData(name string, vector pgvector.Vector, rawData []byte, modelText pgtype.Text, originalContent []byte) database.CreateCreaturesParams {
	size, creatureType, alignment, abilities, skills, speed, languages, senses := parseBestiaryData(originalContent)

	return database.CreateCreaturesParams{
		Name:           name,
		Size:           pgtype.Text{String: size, Valid: true},
		Type:           pgtype.Text{String: creatureType, Valid: creatureType != ""},
		Alignment:      pgtype.Text{String: alignment, Valid: true},
		Abilities:      abilities,
		Skills:         skills,
		Speed:          speed,
		Languages:      pgtype.Text{String: languages, Valid: true},
		Senses:         pgtype.Text{String: senses, Valid: true},
		Embedding:      vector,
		RawData:        rawData,
		EmbeddingModel: modelText,
	}
}

// Simple helpers
func getString(metadata map[string]any, key, defaultValue string) string {
	if val, ok := metadata[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func extractSimpleName(result *EmbeddingResult, content string) string {
	if name := extractFromFilename(result); name != "" {
		return name
	}
	if name := extractFromJSON(content); name != "" {
		return name
	}
	if name := extractFromMarkdown(content); name != "" {
		return name
	}
	return ""
}

func extractFromFilename(result *EmbeddingResult) string {
	filename, ok := result.Metadata["filename"].(string)
	if !ok {
		return ""
	}
	name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	name = strings.ReplaceAll(name, "_", " ")
	return cases.Title(language.English).String(strings.ToLower(name))
}

func extractFromJSON(content string) string {
	if !strings.HasPrefix(content, "{") {
		return ""
	}
	lines := strings.Split(content, "\n")
	limit := min(5, len(lines))
	for _, line := range lines[:limit] {
		if name := parseJSONNameLine(line); name != "" {
			return name
		}
	}
	return ""
}

func parseJSONNameLine(line string) string {
	if !strings.Contains(line, `"name"`) || !strings.Contains(line, ":") {
		return ""
	}
	parts := strings.Split(line, ":")
	if len(parts) <= 1 {
		return ""
	}
	name := strings.Trim(strings.TrimSpace(parts[1]), `,"`)
	return name
}

func extractFromMarkdown(content string) string {
	if !strings.HasPrefix(content, "#") {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.TrimLeft(lines[0], "#"))
}

// parseSpellData extracts spell data from JSON content
func parseSpellData(content []byte) (level int, school, castingTime, rangeValue, components, duration string, classes []string) {
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return getDefaultSpellValues()
	}

	level = extractSpellLevel(data)
	school = getStringFromData(data, "school", defaultNotAvail)
	castingTime = extractCastingTime(data)
	rangeValue = extractRange(data)
	components = extractComponents(data)
	duration = extractDuration(data)
	classes = extractSpellClasses(data)

	return
}

func getDefaultSpellValues() (int, string, string, string, string, string, []string) {
	return 0, defaultNotAvail, defaultNotAvail, defaultNotAvail, defaultNotAvail, defaultNotAvail, []string{}
}

func extractSpellLevel(data map[string]interface{}) int {
	if levelStr, ok := data["level"].(string); ok {
		// Parse "Level 3" format
		if strings.HasPrefix(levelStr, "Level ") {
			if levelNum := strings.TrimPrefix(levelStr, "Level "); levelNum != "" {
				if level, err := strconv.Atoi(levelNum); err == nil {
					return level
				}
			}
		}
		// Handle "Cantrip" or numeric strings
		if levelStr == "Cantrip" {
			return 0
		}
		if level, err := strconv.Atoi(levelStr); err == nil {
			return level
		}
	}
	return 0
}

func extractCastingTime(data map[string]interface{}) string {
	if castingTime, ok := data["casting_time"].(map[string]interface{}); ok {
		if value, valueOk := castingTime["value"].(string); valueOk {
			if unit, unitOk := castingTime["unit"].(string); unitOk {
				return value + " " + unit
			}
		}
	}
	return defaultNotAvail
}

func extractRange(data map[string]interface{}) string {
	if rangeData, ok := data["range"].(map[string]interface{}); ok {
		if value, valueOk := rangeData["value"]; valueOk {
			if unit, unitOk := rangeData["unit"].(string); unitOk {
				return fmt.Sprintf("%v %s", value, unit)
			}
		}
	}
	if rangeStr, ok := data["range"].(string); ok {
		return rangeStr
	}
	return defaultNotAvail
}

func extractComponents(data map[string]interface{}) string {
	if components, ok := data["components"].(map[string]interface{}); ok {
		var parts []string

		if verbal, ok := components["verbal"].(bool); ok && verbal {
			parts = append(parts, "V")
		}

		if somatic, ok := components["somatic"].(bool); ok && somatic {
			parts = append(parts, "S")
		}

		parts = appendMaterialComponent(parts, components)

		if len(parts) > 0 {
			return strings.Join(parts, ", ")
		}
	}
	return defaultNotAvail
}

func appendMaterialComponent(parts []string, components map[string]interface{}) []string {
	if material, ok := components["material"].(string); ok && material != "" {
		return append(parts, "M ("+material+")")
	}
	if hasMaterial, ok := components["material"].(bool); ok && hasMaterial {
		return append(parts, "M")
	}
	return parts
}

func safeIntToInt32(value int) int32 {
	if value > 2147483647 || value < -2147483648 {
		return 0 // Return safe default for spell level
	}
	return int32(value)
}

func extractDuration(data map[string]interface{}) string {
	if duration, ok := data["duration"].(map[string]interface{}); ok {
		if durationType, ok := duration["type"].(string); ok {
			if durationType == "instantaneous" {
				return "Instantaneous"
			}
			if value, valueOk := duration["value"]; valueOk {
				if unit, unitOk := duration["unit"].(string); unitOk {
					return fmt.Sprintf("%v %s", value, unit)
				}
			}
			return durationType
		}
	}
	if duration, ok := data["duration"].(string); ok {
		return duration
	}
	return defaultNotAvail
}

func extractSpellClasses(data map[string]interface{}) []string {
	if classes, ok := data["classes"].([]interface{}); ok {
		var classNames []string
		for _, class := range classes {
			if className, ok := class.(string); ok {
				classNames = append(classNames, className)
			}
		}
		return classNames
	}
	return []string{}
}

// parseBestiaryData extracts creature data from JSON content
func parseBestiaryData(content []byte) (size, creatureType, alignment string, abilities, skills, speed []byte, languages, senses string) {
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return getDefaultCreatureValues()
	}

	size = getStringFromData(data, "size", "Medium")
	creatureType = getStringFromData(data, "type", "")
	alignment = getStringFromData(data, "alignment", "Unaligned")

	abilities = extractAbilities(data)
	skills = extractSkills(data)
	speed = extractSpeed(data)

	languages = extractLanguages(data)
	senses = extractSenses(data)

	return
}

func getDefaultCreatureValues() (string, string, string, []byte, []byte, []byte, string, string) {
	return "Medium", "", "Unaligned", []byte("{}"), []byte("{}"), []byte("{}"), defaultValue, defaultValue
}

func getStringFromData(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func extractAbilities(data map[string]interface{}) []byte {
	if abilities, ok := data["ability_scores"].(map[string]interface{}); ok {
		if jsonData, err := json.Marshal(abilities); err == nil {
			return jsonData
		}
	}
	return []byte("{}")
}

func extractSkills(data map[string]interface{}) []byte {
	if skillsArray, ok := data["skills"].([]interface{}); ok {
		skillsMap := make(map[string]interface{})
		for _, skill := range skillsArray {
			if skillObj, ok := skill.(map[string]interface{}); ok {
				if name, nameOk := skillObj["name"].(string); nameOk {
					skillsMap[name] = skillObj
				}
			}
		}
		if jsonData, err := json.Marshal(skillsMap); err == nil {
			return jsonData
		}
	}
	return []byte("{}")
}

func extractSpeed(data map[string]interface{}) []byte {
	if speed, ok := data["speed"].(map[string]interface{}); ok {
		if jsonData, err := json.Marshal(speed); err == nil {
			return jsonData
		}
	}
	return []byte("{}")
}

const (
	defaultValue    = "—"
	defaultNotAvail = "N/A"
)

func extractLanguages(data map[string]interface{}) string {
	if languages, ok := data["languages"].([]interface{}); ok {
		var langStrings []string
		for _, lang := range languages {
			if langStr, ok := lang.(string); ok {
				langStrings = append(langStrings, langStr)
			}
		}
		if len(langStrings) > 0 {
			return strings.Join(langStrings, ", ")
		}
	}
	return defaultValue
}

func extractSenses(data map[string]interface{}) string {
	if senses, ok := data["senses"].(map[string]interface{}); ok {
		var senseStrings []string

		for senseType, senseData := range senses {
			if senseType == "passive_perception" {
				if pp, ok := senseData.(float64); ok {
					senseStrings = append(senseStrings, fmt.Sprintf("passive Perception %d", int(pp)))
				}
			} else {
				if senseMap, ok := senseData.(map[string]interface{}); ok {
					if value, valueOk := senseMap["value"].(float64); valueOk {
						if unit, unitOk := senseMap["unit"].(string); unitOk {
							senseStrings = append(senseStrings, fmt.Sprintf("%s %d %s", senseType, int(value), unit))
						}
					}
				}
			}
		}

		if len(senseStrings) > 0 {
			return strings.Join(senseStrings, ", ")
		}
	}
	return defaultValue
}

// ensureValidJSON converts content to valid JSON for database storage.
// If content is already valid JSON, returns it as-is.
// If not, wraps it in a JSON object with a "content" field.
func ensureValidJSON(content []byte) ([]byte, error) {
	// First check if it's already valid JSON
	var dummy interface{}
	if err := json.Unmarshal(content, &dummy); err == nil {
		return content, nil
	}

	// If not valid JSON, wrap it in a JSON object
	wrapper := map[string]string{
		"content": string(content),
	}
	return json.Marshal(wrapper)
}
