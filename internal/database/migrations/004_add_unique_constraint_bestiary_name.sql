-- +goose Up
-- Add unique constraint for bestiary name to prevent duplicates

ALTER TABLE scry_quest.bestiary
ADD CONSTRAINT idx_unique_bestiary_name UNIQUE (name);

-- +goose Down
-- Remove unique constraint for bestiary name

ALTER TABLE scry_quest.bestiary
DROP CONSTRAINT IF EXISTS idx_unique_bestiary_name;