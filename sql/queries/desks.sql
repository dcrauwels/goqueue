-- name: CreateDesks :one
INSERT INTO desks (id, public_id, name, description, is_active)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    TRUE
)
RETURNING *;

-- name: GetDesksByPublicID :one
SELECT * FROM desks
WHERE public_id = $1;

-- name: GetDesks :many
SELECT * FROM desks;

-- name: GetActiveDesks :many
SELECT * FROM desks
WHERE is_active = TRUE;

-- name: SetDesksByPublicID :one
UPDATE desks
SET name = $2, description = $3, is_active = $4, updated_at = NOW()
WHERE public_id = $1
RETURNING *;

-- name: ListDesks :many
SELECT * FROM desks
WHERE (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.narg('is_active'))
ORDER BY name ASC;