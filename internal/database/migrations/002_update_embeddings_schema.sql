-- +goose Up
-- Update embeddings schema for Ollama models
-- This migration updates the vector dimensions and adds metadata for embedding models

-- Drop existing indexes to allow column updates
DROP INDEX IF EXISTS idx_spells_embedding;
DROP INDEX IF EXISTS idx_bestiary_embedding;
DROP INDEX IF EXISTS idx_classes_embedding;
DROP INDEX IF EXISTS idx_species_embedding;

-- Add model metadata columns to track which embedding model was used
ALTER TABLE scry_quest.spells ADD COLUMN IF NOT EXISTS embedding_model TEXT DEFAULT 'nomic-embed-text';
ALTER TABLE scry_quest.bestiary ADD COLUMN IF NOT EXISTS embedding_model TEXT DEFAULT 'nomic-embed-text';
ALTER TABLE scry_quest.classes ADD COLUMN IF NOT EXISTS embedding_model TEXT DEFAULT 'nomic-embed-text';
ALTER TABLE scry_quest.species ADD COLUMN IF NOT EXISTS embedding_model TEXT DEFAULT 'nomic-embed-text';

-- Update embedding column dimensions for nomic-embed-text (768 dimensions)
ALTER TABLE scry_quest.spells ALTER COLUMN embedding TYPE VECTOR(768);
ALTER TABLE scry_quest.bestiary ALTER COLUMN embedding TYPE VECTOR(768);
ALTER TABLE scry_quest.classes ALTER COLUMN embedding TYPE VECTOR(768);
ALTER TABLE scry_quest.species ALTER COLUMN embedding TYPE VECTOR(768);

-- Update comments to reflect Ollama models
COMMENT ON COLUMN scry_quest.spells.embedding IS 'Embedding vector from Ollama model (default: nomic-embed-text 768 dims)';
COMMENT ON COLUMN scry_quest.bestiary.embedding IS 'Embedding vector from Ollama model (default: nomic-embed-text 768 dims)';
COMMENT ON COLUMN scry_quest.classes.embedding IS 'Embedding vector from Ollama model (default: nomic-embed-text 768 dims)';
COMMENT ON COLUMN scry_quest.species.embedding IS 'Embedding vector from Ollama model (default: nomic-embed-text 768 dims)';

-- Recreate indexes for vector similarity search with updated dimensions
-- Using ivfflat for approximate nearest neighbor search
CREATE INDEX IF NOT EXISTS idx_spells_embedding ON scry_quest.spells USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_bestiary_embedding ON scry_quest.bestiary USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_classes_embedding ON scry_quest.classes USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_species_embedding ON scry_quest.species USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Add additional indexes for embedding model tracking
CREATE INDEX IF NOT EXISTS idx_spells_embedding_model ON scry_quest.spells(embedding_model);
CREATE INDEX IF NOT EXISTS idx_bestiary_embedding_model ON scry_quest.bestiary(embedding_model);
CREATE INDEX IF NOT EXISTS idx_classes_embedding_model ON scry_quest.classes(embedding_model);
CREATE INDEX IF NOT EXISTS idx_species_embedding_model ON scry_quest.species(embedding_model);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION validate_embedding_dimensions()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if the embedding dimension matches the expected model dimension
    CASE NEW.embedding_model
        WHEN 'gpt-oss:20b' THEN
            IF vector_dims(NEW.embedding) != 2880 THEN
                RAISE EXCEPTION 'Embedding dimension % does not match expected 2880 for model %', 
                    vector_dims(NEW.embedding), NEW.embedding_model;
            END IF;
        WHEN 'nomic-embed-text' THEN
            IF vector_dims(NEW.embedding) != 768 THEN
                RAISE EXCEPTION 'Embedding dimension % does not match expected 768 for model %', 
                    vector_dims(NEW.embedding), NEW.embedding_model;
            END IF;
        WHEN 'all-minilm' THEN
            IF vector_dims(NEW.embedding) != 384 THEN
                RAISE EXCEPTION 'Embedding dimension % does not match expected 384 for model %', 
                    vector_dims(NEW.embedding), NEW.embedding_model;
            END IF;
        ELSE
            -- For unknown models, just log a warning
            RAISE NOTICE 'Unknown embedding model: %', NEW.embedding_model;
    END CASE;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Add validation triggers
DROP TRIGGER IF EXISTS validate_spells_embedding_dimensions ON scry_quest.spells;
CREATE TRIGGER validate_spells_embedding_dimensions 
    BEFORE INSERT OR UPDATE ON scry_quest.spells 
    FOR EACH ROW 
    WHEN (NEW.embedding IS NOT NULL AND NEW.embedding_model IS NOT NULL)
    EXECUTE FUNCTION validate_embedding_dimensions();

DROP TRIGGER IF EXISTS validate_bestiary_embedding_dimensions ON scry_quest.bestiary;
CREATE TRIGGER validate_bestiary_embedding_dimensions 
    BEFORE INSERT OR UPDATE ON scry_quest.bestiary 
    FOR EACH ROW 
    WHEN (NEW.embedding IS NOT NULL AND NEW.embedding_model IS NOT NULL)
    EXECUTE FUNCTION validate_embedding_dimensions();

DROP TRIGGER IF EXISTS validate_classes_embedding_dimensions ON scry_quest.classes;
CREATE TRIGGER validate_classes_embedding_dimensions 
    BEFORE INSERT OR UPDATE ON scry_quest.classes 
    FOR EACH ROW 
    WHEN (NEW.embedding IS NOT NULL AND NEW.embedding_model IS NOT NULL)
    EXECUTE FUNCTION validate_embedding_dimensions();

DROP TRIGGER IF EXISTS validate_species_embedding_dimensions ON scry_quest.species;
CREATE TRIGGER validate_species_embedding_dimensions 
    BEFORE INSERT OR UPDATE ON scry_quest.species 
    FOR EACH ROW 
    WHEN (NEW.embedding IS NOT NULL AND NEW.embedding_model IS NOT NULL)
    EXECUTE FUNCTION validate_embedding_dimensions();

-- Add some useful views for embedding statistics
CREATE OR REPLACE VIEW scry_quest.embedding_stats AS
SELECT 
    'spells' as table_name,
    COUNT(*) as total_rows,
    COUNT(embedding) as embedded_rows,
    embedding_model,
    CASE embedding_model
        WHEN 'gpt-oss:20b' THEN 2880
        WHEN 'nomic-embed-text' THEN 768
        WHEN 'all-minilm' THEN 384
        ELSE NULL
    END as expected_dimensions
FROM scry_quest.spells
GROUP BY embedding_model

UNION ALL

SELECT 
    'bestiary' as table_name,
    COUNT(*) as total_rows,
    COUNT(embedding) as embedded_rows,
    embedding_model,
    CASE embedding_model
        WHEN 'gpt-oss:20b' THEN 2880
        WHEN 'nomic-embed-text' THEN 768
        WHEN 'all-minilm' THEN 384
        ELSE NULL
    END as expected_dimensions
FROM scry_quest.bestiary
GROUP BY embedding_model

UNION ALL

SELECT 
    'classes' as table_name,
    COUNT(*) as total_rows,
    COUNT(embedding) as embedded_rows,
    embedding_model,
    CASE embedding_model
        WHEN 'gpt-oss:20b' THEN 2880
        WHEN 'nomic-embed-text' THEN 768
        WHEN 'all-minilm' THEN 384
        ELSE NULL
    END as expected_dimensions
FROM scry_quest.classes
GROUP BY embedding_model

UNION ALL

SELECT 
    'species' as table_name,
    COUNT(*) as total_rows,
    COUNT(embedding) as embedded_rows,
    embedding_model,
    CASE embedding_model
        WHEN 'gpt-oss:20b' THEN 2880
        WHEN 'nomic-embed-text' THEN 768
        WHEN 'all-minilm' THEN 384
        ELSE NULL
    END as expected_dimensions
FROM scry_quest.species
GROUP BY embedding_model;

-- +goose Down
-- Remove the embedding statistics view
DROP VIEW IF EXISTS scry_quest.embedding_stats;

-- Remove validation triggers and function
DROP TRIGGER IF EXISTS validate_spells_embedding_dimensions ON scry_quest.spells;
DROP TRIGGER IF EXISTS validate_bestiary_embedding_dimensions ON scry_quest.bestiary;
DROP TRIGGER IF EXISTS validate_classes_embedding_dimensions ON scry_quest.classes;
DROP TRIGGER IF EXISTS validate_species_embedding_dimensions ON scry_quest.species;
DROP FUNCTION IF EXISTS validate_embedding_dimensions;

-- Remove embedding model indexes
DROP INDEX IF EXISTS idx_spells_embedding_model;
DROP INDEX IF EXISTS idx_bestiary_embedding_model;  
DROP INDEX IF EXISTS idx_classes_embedding_model;
DROP INDEX IF EXISTS idx_species_embedding_model;

-- Remove embedding model columns
ALTER TABLE scry_quest.spells DROP COLUMN IF EXISTS embedding_model;
ALTER TABLE scry_quest.bestiary DROP COLUMN IF EXISTS embedding_model;
ALTER TABLE scry_quest.classes DROP COLUMN IF EXISTS embedding_model;
ALTER TABLE scry_quest.species DROP COLUMN IF EXISTS embedding_model;