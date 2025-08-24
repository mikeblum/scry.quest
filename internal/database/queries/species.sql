-- name: CreateSpecies :one
INSERT INTO scry_quest.species (name, description, size, speed, ability_score_increase, traits, embedding, embedding_model, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetSpeciesByID :one
SELECT * FROM scry_quest.species WHERE id = $1;

-- name: GetSpeciesByName :one
SELECT * FROM scry_quest.species WHERE name = $1;

-- name: ListSpecies :many
SELECT * FROM scry_quest.species
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: SearchSpeciesByEmbedding :many
SELECT 
    id,
    name,
    description,
    size,
    speed,
    embedding_model,
    (1 - (embedding <=> $1)) as similarity
FROM scry_quest.species
WHERE embedding IS NOT NULL
ORDER BY embedding <=> $1
LIMIT $2;

-- name: UpdateSpeciesEmbedding :exec
UPDATE scry_quest.species 
SET embedding = $2, embedding_model = $3, updated_at = NOW()
WHERE id = $1;

-- name: DeleteSpecies :exec
DELETE FROM scry_quest.species WHERE id = $1;