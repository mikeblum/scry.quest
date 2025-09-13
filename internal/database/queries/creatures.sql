-- name: CreateCreature :one
INSERT INTO scry_quest.bestiary (name, size, type, subtype, alignment, armor_class, hit_points, hit_dice, speed, abilities, skills, senses, languages, challenge_rating, embedding, embedding_model, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: CreateCreatures :batchexec
INSERT INTO scry_quest.bestiary (name, size, type, subtype, alignment, armor_class, hit_points, hit_dice, speed, abilities, skills, senses, languages, challenge_rating, embedding, embedding_model, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
ON CONFLICT (name) DO UPDATE SET
    size = EXCLUDED.size,
    type = EXCLUDED.type,
    subtype = EXCLUDED.subtype,
    alignment = EXCLUDED.alignment,
    armor_class = EXCLUDED.armor_class,
    hit_points = EXCLUDED.hit_points,
    hit_dice = EXCLUDED.hit_dice,
    speed = EXCLUDED.speed,
    abilities = EXCLUDED.abilities,
    skills = EXCLUDED.skills,
    senses = EXCLUDED.senses,
    languages = EXCLUDED.languages,
    challenge_rating = EXCLUDED.challenge_rating,
    embedding = EXCLUDED.embedding,
    embedding_model = EXCLUDED.embedding_model,
    raw_data = EXCLUDED.raw_data,
    updated_at = NOW();

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

-- name: SearchCreaturesByEmbedding :many
SELECT 
    id,
    name,
    size,
    type,
    alignment,
    challenge_rating,
    embedding_model,
    (1 - (embedding <=> $1))::float8 as similarity
FROM scry_quest.bestiary
WHERE embedding IS NOT NULL
ORDER BY embedding <=> $1
LIMIT $2;

-- name: UpdateCreatureEmbedding :exec
UPDATE scry_quest.bestiary 
SET embedding = $2, embedding_model = $3, updated_at = NOW()
WHERE id = $1;

-- name: DeleteBeast :exec
DELETE FROM scry_quest.bestiary WHERE id = $1;