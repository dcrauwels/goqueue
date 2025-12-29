-- name: CreateVisitor :one
INSERT INTO visitors (id, public_id, created_at, updated_at, waiting_since, name, purpose_public_id, status, daily_ticket_number)
VALUES (
    gen_random_uuid(),
    $1,
    NOW(),
    NOW(),
    NOW(),
    $2,
    $3,
    0, --status 
    $4
)
RETURNING *;

-- name: GetVisitors :many
SELECT * FROM visitors;

-- name: GetVisitorsByPublicID :one
SELECT * FROM visitors
WHERE public_id = $1;

-- name: GetVisitorByID :one
SELECT * FROM visitors
WHERE visitors.id = $1;

-- name: GetVisitorsByStatus :many
SELECT * FROM visitors
WHERE status = $1 -- status
ORDER BY waiting_since ASC;

-- name: SetVisitorByPublicID :one
UPDATE visitors
SET name = $2, purpose_public_id = $3, status = $4, updated_at = NOW() -- status
WHERE public_id = $1
RETURNING *;

-- name: GetVisitorsByPurposePublicID :many
SELECT * FROM visitors
WHERE purpose_public_id = $1
ORDER BY waiting_since ASC;

-- name: GetVisitorsByPurposePublicIDAndStatus :many
SELECT * FROM visitors
WHERE purpose_public_id = $1 AND status = $2
ORDER BY waiting_since ASC;

-- name: GetWaitingVisitorsByPurposePublicID :many
SELECT * FROM visitors 
WHERE purpose_public_id = $1 AND status = 1 -- NOTE that statuses are still not properly implemented
ORDER BY waiting_since ASC;

-- name: GetVisitorsForToday :many
SELECT * FROM visitors
WHERE waiting_since::date = CURRENT_DATE
ORDER BY waiting_since ASC;

-- name: SetVisitorStatusByID :one
UPDATE visitors
SET status = $2, updated_at = NOW() --status 
WHERE id = $1
RETURNING *;

-- name: ListVisitors :many
SELECT * FROM visitors
WHERE (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
    AND (sqlc.narg('purpose_public_id')::text IS NULL OR purpose_public_id = sqlc.narg('purpose_public_id'))
    AND (sqlc.narg('start_date')::timestamp IS NULL OR created_at >= sqlc.narg('start_date'))
    AND (sqlc.narg('end_date')::timestamp IS NULL OR created_at < sqlc.narg('end_date'))
ORDER BY waiting_since ASC;