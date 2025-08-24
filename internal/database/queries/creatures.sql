-- name: CreateBeast :one
INSERT INTO scry_quest.bestiary (name, size, type, subtype, alignment, armor_class, hit_points, hit_dice, speed, abilities, skills, senses, languages, challenge_rating, embedding, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: GetBeastByID :one
SELECT * FROM scry_quest.bestiary WHERE id = $1;

-- name: GetBeastByName :one
SELECT * FROM scry_quest.bestiary WHERE name = $1;

-- name: ListBeasts :many
SELECT * FROM scry_quest.bestiary
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: ListBeastsByType :many
SELECT * FROM scry_quest.bestiary
WHERE type = $1
ORDER BY name;

-- name: ListBeastsByChallengeRating :many
SELECT * FROM scry_quest.bestiary
WHERE challenge_rating = $1
ORDER BY name;

-- name: SearchBeastsByEmbedding :many
SELECT *, (embedding <=> $1::vector) as distance
FROM scry_quest.bestiary
ORDER BY embedding <=> $1::vector
LIMIT $2;

-- name: UpdateBeastEmbedding :exec
UPDATE scry_quest.bestiary 
SET embedding = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteBeast :exec
DELETE FROM scry_quest.bestiary WHERE id = $1;