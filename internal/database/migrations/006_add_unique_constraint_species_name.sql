-- +goose Up
-- Add unique constraint for species name to prevent duplicates

ALTER TABLE scry_quest.species
ADD CONSTRAINT idx_unique_species_name UNIQUE (name);

-- +goose Down
-- Remove unique constraint for species name

ALTER TABLE scry_quest.species
DROP CONSTRAINT IF EXISTS idx_unique_species_name;