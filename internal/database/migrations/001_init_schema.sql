-- Enable pgvector extension for embeddings
CREATE EXTENSION IF NOT EXISTS vector;

-- Create scry_quest schema and set ownership
CREATE SCHEMA IF NOT EXISTS scry_quest AUTHORIZATION scry_quest;

-- Set default privileges for future objects created by scry_quest user
ALTER DEFAULT PRIVILEGES FOR USER scry_quest IN SCHEMA scry_quest GRANT ALL ON TABLES TO scry_quest;
ALTER DEFAULT PRIVILEGES FOR USER scry_quest IN SCHEMA scry_quest GRANT ALL ON SEQUENCES TO scry_quest;
ALTER DEFAULT PRIVILEGES FOR USER scry_quest IN SCHEMA scry_quest GRANT ALL ON FUNCTIONS TO scry_quest;

-- Set default schema for the user
ALTER USER scry_quest SET search_path = scry_quest, public;

-- Table for storing D&D spells with embeddings
CREATE TABLE scry_quest.spells (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    level INTEGER NOT NULL,
    school TEXT,
    casting_time TEXT,
    range_value TEXT,
    components TEXT,
    duration TEXT,
    classes TEXT[],
    embedding VECTOR(1536), -- OpenAI gpt-oss-20b dimension -- gpt-oss-20b dimensions
    raw_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NULL
);

-- Table for storing D&D bestiary with embeddings  
CREATE TABLE scry_quest.bestiary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    size TEXT,
    type TEXT,
    subtype TEXT,
    alignment TEXT,
    armor_class INTEGER,
    hit_points INTEGER,
    hit_dice TEXT,
    speed JSONB,
    abilities JSONB, -- STR, DEX, CON, INT, WIS, CHA
    skills JSONB,
    senses TEXT,
    languages TEXT,
    challenge_rating TEXT,
    embedding VECTOR(1536), -- OpenAI gpt-oss-20b dimension
    raw_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NULL
);

-- Table for storing D&D classes with embeddings
CREATE TABLE scry_quest.classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    hit_die INTEGER,
    primary_ability TEXT,
    saving_throw_proficiencies TEXT,
    skill_proficiencies TEXT[],
    embedding VECTOR(1536), -- OpenAI gpt-oss-20b dimension
    raw_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NULL
);

-- Table for storing D&D species/races with embeddings
CREATE TABLE scry_quest.species (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    size TEXT,
    speed INTEGER,
    ability_score_increase JSONB,
    traits TEXT[],
    embedding VECTOR(1536), -- OpenAI gpt-oss-20b dimension
    raw_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NULL
);

-- Indexes for vector similarity search
CREATE INDEX idx_spells_embedding ON scry_quest.spells USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_bestiary_embedding ON scry_quest.bestiary USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_classes_embedding ON scry_quest.classes USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX idx_species_embedding ON scry_quest.species USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Regular indexes for common queries
CREATE INDEX idx_spells_name ON scry_quest.spells(name);
CREATE INDEX idx_spells_level ON scry_quest.spells(level);
CREATE INDEX idx_spells_school ON scry_quest.spells(school);
CREATE INDEX idx_bestiary_name ON scry_quest.bestiary(name);
CREATE INDEX idx_bestiary_type ON scry_quest.bestiary(type);
CREATE INDEX idx_bestiary_challenge_rating ON scry_quest.bestiary(challenge_rating);
CREATE INDEX idx_classes_name ON scry_quest.classes(name);
CREATE INDEX idx_species_name ON scry_quest.species(name);

-- Update trigger function for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers to all tables
CREATE TRIGGER update_spells_updated_at BEFORE UPDATE ON scry_quest.spells FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_bestiary_updated_at BEFORE UPDATE ON scry_quest.bestiary FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_classes_updated_at BEFORE UPDATE ON scry_quest.classes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_species_updated_at BEFORE UPDATE ON scry_quest.species FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();