-- +goose Up
-- Add unique constraint for spells name to prevent duplicates

ALTER TABLE scry_quest.spells
ADD CONSTRAINT idx_unique_spells_name UNIQUE (name);

-- +goose Down
-- Remove unique constraint for spells name

ALTER TABLE scry_quest.spells
DROP CONSTRAINT IF EXISTS idx_unique_spells_name;