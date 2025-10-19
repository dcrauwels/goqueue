-- name: CreateDesks :one
INSERT INTO desks (id, public_id, number, description, is_active)
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