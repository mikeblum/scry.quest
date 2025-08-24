-- name: CreateClass :one
INSERT INTO scry_quest.classes (name, description, hit_die, primary_ability, saving_throw_proficiencies, skill_proficiencies, embedding, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetClassByID :one
SELECT * FROM scry_quest.classes WHERE id = $1;

-- name: GetClassByName :one
SELECT * FROM scry_quest.classes WHERE name = $1;

-- name: ListClasses :many
SELECT * FROM scry_quest.classes
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: SearchClassesByEmbedding :many
SELECT *, (embedding <=> $1::vector) as distance
FROM scry_quest.classes
ORDER BY embedding <=> $1::vector
LIMIT $2;

-- name: UpdateClassEmbedding :exec
UPDATE scry_quest.classes 
SET embedding = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteClass :exec
DELETE FROM scry_quest.classes WHERE id = $1;