-- name: CreateVisitor :one
INSERT INTO visitors (id, created_at, updated_at, waiting_since, name, purpose_id, status)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    NOW(),
    $1,
    $2,
    0 --status 
)
RETURNING *;

-- name: GetVisitors :many
SELECT * FROM visitors;


-- name: GetVisitorByID :one
SELECT * FROM visitors
WHERE visitors.id = $1;

-- name: GetVisitorsByStatus :many
SELECT * FROM visitors
WHERE status = $1 -- status
ORDER BY waiting_since ASC;

-- name: GetVisitorsByPurpose :many
SELECT * FROM visitors
WHERE purpose_id = $1
ORDER BY waiting_since ASC;

-- name: GetVisitorsByPurposeStatus :many
SELECT * FROM visitors
WHERE purpose_id = $1 AND status = $2
ORDER BY waiting_since ASC;

-- name: GetWaitingVisitorsByPurpose :many
SELECT * FROM visitors
WHERE purpose_id = $1 AND status = 1 -- this whole status business is still not implemented correctly
ORDER BY waiting_since ASC;

-- name: GetVisitorsForToday :many
SELECT * FROM visitors
WHERE waiting_since::date = CURRENT_DATE
ORDER BY waiting_since ASC;

-- name: SetVisitorByID :one
UPDATE visitors
SET name = $2, purpose_id = $3, status = $4, updated_at = NOW() -- status 
WHERE id = $1
RETURNING *;

-- name: SetVisitorStatusByID :one
UPDATE visitors
SET status = $2, updated_at = NOW() --status 
WHERE id = $1
RETURNING *;