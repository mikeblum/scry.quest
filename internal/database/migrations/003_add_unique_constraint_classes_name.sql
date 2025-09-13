-- +goose Up
-- Add unique constraint for class name to prevent duplicates

ALTER TABLE scry_quest.classes
ADD CONSTRAINT idx_unique_classes_name UNIQUE (name);

-- +goose Down
-- Remove unique constraint for class name

ALTER TABLE scry_quest.classes
DROP CONSTRAINT IF EXISTS idx_unique_classes_name;