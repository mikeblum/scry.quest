
-- name: SearchAllContentByEmbedding :many
SELECT 
    'spell' as content_type,
    s.id::text,
    s.name,
    s.description,
    s.embedding_model,
    (1 - (s.embedding <=> $1))::float8 as similarity
FROM scry_quest.spells s
WHERE s.embedding IS NOT NULL

UNION ALL

SELECT 
    'creature' as content_type,
    b.id::text,
    b.name,
    CONCAT(b.size, ' ', b.type, ', ', b.alignment) as description,
    b.embedding_model,
    (1 - (b.embedding <=> $1))::float8 as similarity
FROM scry_quest.bestiary b
WHERE b.embedding IS NOT NULL

UNION ALL

SELECT 
    'class' as content_type,
    c.id::text,
    c.name,
    c.description,
    c.embedding_model,
    (1 - (c.embedding <=> $1))::float8 as similarity
FROM scry_quest.classes c
WHERE c.embedding IS NOT NULL

UNION ALL

SELECT 
    'species' as content_type,
    sp.id::text,
    sp.name,
    sp.description,
    sp.embedding_model,
    (1 - (sp.embedding <=> $1))::float8 as similarity
FROM scry_quest.species sp
WHERE sp.embedding IS NOT NULL

ORDER BY similarity DESC
LIMIT $2;

-- name: GetEmbeddingStats :many
SELECT * FROM scry_quest.embedding_stats ORDER BY table_name, embedding_model;

-- name: DeleteSpellEmbeddings :exec
UPDATE scry_quest.spells 
SET embedding = NULL, embedding_model = NULL 
WHERE embedding_model = $1 OR $1 = 'all';

-- name: DeleteCreatureEmbeddings :exec
UPDATE scry_quest.bestiary 
SET embedding = NULL, embedding_model = NULL 
WHERE embedding_model = $1 OR $1 = 'all';

-- name: DeleteClassEmbeddings :exec
UPDATE scry_quest.classes 
SET embedding = NULL, embedding_model = NULL 
WHERE embedding_model = $1 OR $1 = 'all';

-- name: DeleteSpeciesEmbeddings :exec
UPDATE scry_quest.species 
SET embedding = NULL, embedding_model = NULL 
WHERE embedding_model = $1 OR $1 = 'all';


-- name: GetSpellsWithoutEmbeddings :many
SELECT id, name, description, level, school, raw_data
FROM scry_quest.spells
WHERE embedding IS NULL
ORDER BY created_at DESC;

-- name: GetCreaturesWithoutEmbeddings :many
SELECT id, name, size, type, alignment, challenge_rating, raw_data
FROM scry_quest.bestiary
WHERE embedding IS NULL
ORDER BY created_at DESC;

-- name: GetClassesWithoutEmbeddings :many
SELECT id, name, description, raw_data
FROM scry_quest.classes
WHERE embedding IS NULL
ORDER BY created_at DESC;

-- name: GetSpeciesWithoutEmbeddings :many
SELECT id, name, description, raw_data
FROM scry_quest.species
WHERE embedding IS NULL
ORDER BY created_at DESC;

-- name: CountItemsByEmbeddingModel :many
SELECT 
    'spells' as table_name,
    COALESCE(embedding_model, 'no_embedding') as model,
    COUNT(*) as count
FROM scry_quest.spells
GROUP BY embedding_model

UNION ALL

SELECT 
    'bestiary' as table_name,
    COALESCE(embedding_model, 'no_embedding') as model,
    COUNT(*) as count
FROM scry_quest.bestiary
GROUP BY embedding_model

UNION ALL

SELECT 
    'classes' as table_name,
    COALESCE(embedding_model, 'no_embedding') as model,
    COUNT(*) as count
FROM scry_quest.classes
GROUP BY embedding_model

UNION ALL

SELECT 
    'species' as table_name,
    COALESCE(embedding_model, 'no_embedding') as model,
    COUNT(*) as count
FROM scry_quest.species
GROUP BY embedding_model

ORDER BY table_name, model;