-- name: CreateClass :one
INSERT INTO scry_quest.classes (name, description, hit_die, primary_ability, saving_throw_proficiencies, skill_proficiencies, embedding, embedding_model, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: CreateClasses :batchexec
INSERT INTO scry_quest.classes (name, description, hit_die, primary_ability, saving_throw_proficiencies, skill_proficiencies, embedding, embedding_model, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    hit_die = EXCLUDED.hit_die,
    primary_ability = EXCLUDED.primary_ability,
    saving_throw_proficiencies = EXCLUDED.saving_throw_proficiencies,
    skill_proficiencies = EXCLUDED.skill_proficiencies,
    embedding = EXCLUDED.embedding,
    embedding_model = EXCLUDED.embedding_model,
    raw_data = EXCLUDED.raw_data,
    updated_at = NOW();

-- name: GetClassByID :one
SELECT * FROM scry_quest.classes WHERE id = $1;

-- name: GetClassByName :one
SELECT * FROM scry_quest.classes WHERE name = $1;

-- name: ListClasses :many
SELECT * FROM scry_quest.classes
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: SearchClassesByEmbedding :many
SELECT 
    id,
    name,
    description,
    primary_ability,
    embedding_model,
    (1 - (embedding <=> $1))::float8 as similarity
FROM scry_quest.classes
WHERE embedding IS NOT NULL
ORDER BY embedding <=> $1
LIMIT $2;

-- name: UpdateClassEmbedding :exec
UPDATE scry_quest.classes 
SET embedding = $2, embedding_model = $3, updated_at = NOW()
WHERE id = $1;

-- name: DeleteClass :exec
DELETE FROM scry_quest.classes WHERE id = $1;