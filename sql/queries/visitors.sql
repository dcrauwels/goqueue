-- name: CreateVisitor :one
INSERT INTO visitors (id, created_at, updated_at, waiting_since, name, purpose, status)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    NOW(),
    $1,
    $2,
    0
)
RETURNING *;

-- name: GetVisitorByID :one
SELECT * FROM visitors
WHERE id = $1;

-- name: GetVisitorsByStatus :many
SELECT * FROM visitors
WHERE status = $1
ORDER BY waiting_since ASC;

-- name: GetWaitingVisitorsByPurpose :many
SELECT * FROM visitors
WHERE purpose = $1 AND status = 1
ORDER BY waiting_since ASC;