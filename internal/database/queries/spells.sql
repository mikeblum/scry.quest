-- name: CreateSpell :one
INSERT INTO scry_quest.spells (name, description, level, school, casting_time, range_value, components, duration, classes, embedding, embedding_model, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetSpellByID :one
SELECT * FROM scry_quest.spells WHERE id = $1;

-- name: GetSpellByName :one
SELECT * FROM scry_quest.spells WHERE name = $1;

-- name: ListSpells :many
SELECT * FROM scry_quest.spells
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: ListSpellsByLevel :many
SELECT * FROM scry_quest.spells
WHERE level = $1
ORDER BY name;

-- name: ListSpellsBySchool :many
SELECT * FROM scry_quest.spells
WHERE school = $1
ORDER BY name;

-- name: SearchSpellsByEmbedding :many
SELECT 
    id,
    name,
    description,
    level,
    school,
    embedding_model,
    (1 - (embedding <=> $1)) as similarity
FROM scry_quest.spells
WHERE embedding IS NOT NULL
ORDER BY embedding <=> $1
LIMIT $2;

-- name: UpdateSpellEmbedding :exec
UPDATE scry_quest.spells 
SET embedding = $2, embedding_model = $3, updated_at = NOW()
WHERE id = $1;

-- name: DeleteSpell :exec
DELETE FROM scry_quest.spells WHERE id = $1;