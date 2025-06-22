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

-- name: GetVisitorsForToday :many
SELECT * FROM visitors
WHERE waiting_since::date = CURRENT_DATE
ORDER BY waiting_since ASC;

-- name: SetVisitorByID :one
UPDATE visitors
SET name = $2, purpose = $3, status = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetVisitorStatusByID :one
UPDATE visitors
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;