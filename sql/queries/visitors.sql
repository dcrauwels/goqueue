-- name: CreateVisitor :one
WITH inserted_visitor AS (
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
    RETURNING *
)
SELECT iv.public_id, iv.waiting_since, iv.name, iv.status, iv.daily_ticket_number, iv.purpose_public_id, p.purpose_name AS purpose_name
FROM inserted_visitor iv
INNER JOIN purposes p 
ON p.public_id = iv.purpose_public_id;

-- name: GetVisitors :many
SELECT * FROM visitor_response_values;

-- name: GetVisitorsByPublicID :one
SELECT * FROM visitor_response_values
WHERE public_id = $1;

-- name: GetVisitorsByStatus :many
SELECT * FROM visitor_response_values
WHERE status = $1
ORDER BY waiting_since ASC;

-- name: SetVisitorByPublicID :one
WITH updated_visitor AS(
    UPDATE visitors
    SET name = $2, purpose_public_id = $3, status = $4, updated_at = NOW()
    WHERE visitors.public_id = $1
    RETURNING *
)
SELECT uv.public_id, uv.waiting_since, uv.name, uv.status, uv.daily_ticket_number, uv.purpose_public_id, p.purpose_name
FROM updated_visitor uv
INNER JOIN purposes p
ON p.public_id = uv.purpose_public_id;

-- name: GetVisitorsByPurposePublicID :many
SELECT * FROM visitor_response_values
WHERE purpose_public_id = $1
ORDER BY waiting_since ASC;

-- name: GetVisitorsByPurposePublicIDAndStatus :many
SELECT * FROM visitor_response_values
WHERE purpose_public_id $1 AND status = $2
ORDER BY waiting_since ASC;

-- name: GetWaitingVisitorsByPurposePublicID :many
SELECT * FROM visitor_response_values
WHERE purpose_public_id $1 AND status = 1
ORDER BY waiting_since ASC;

-- name: GetVisitorsForToday :many
SELECT * FROM visitor_response_values
WHERE waiting_since::date = CURRENT_DATE
ORDER BY waiting_since ASC;

-- name: SetVisitorStatusByPublicID :one
WITH updated_visitor AS (
    UPDATE visitors
    SET status = $2, updated_at = NOW() --status 
    WHERE visitors.public_id = $1
    RETURNING *
)
SELECT uv.public_id, uv.waiting_since, uv.name, uv.status, uv.daily_ticket_number, uv.purpose_public_id, p.purpose_name
FROM updated_visitor uv
INNER JOIN purposes p
ON p.public_id = uv.purpose_public_id;

-- name: ListVisitors :many
SELECT * FROM visitor_response_values
WHERE (sqlc.narg('status')::int IS NULL OR status = sqlc.narg('status'))
    AND (sqlc.narg('purpose_public_id')::text IS NULL OR purpose_public_id = sqlc.narg('purpose_public_id'))
    AND (sqlc.narg('start_date')::timestamp IS NULL OR waiting_since >= sqlc.narg('start_date'))
    AND (sqlc.narg('end_date')::timestamp IS NULL OR waiting_since < sqlc.narg('end_date'))
ORDER BY waiting_since ASC;

-- name: GetQueue :many
SELECT * FROM visitor_response_values
WHERE status IN (1,2,3);

-- name: GetPublicIDOfNextWaitingVisitor :one
SELECT public_id FROM visitor_response_values
WHERE status = 1
ORDER BY waiting_since ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: CallVisitorByPublicID :one
WITH updated_visitor AS (
    UPDATE visitors
    SET status = 2, updated_at = NOW()
    WHERE visitors.public_id = $1
    RETURNING *
)
SELECT uv.public_id, uv.waiting_since, uv.name, uv.status, uv.daily_ticket_number, uv.purpose_public_id, p.purpose_name
FROM updated_visitor uv
INNER JOIN purposes p
ON p.public_id = uv.purpose_public_id;

-- name: CallNextWaitingVisitor :one
WITH next_visitor_public_id AS (
    SELECT public_id FROM visitor_response_values
    WHERE status = 1
    ORDER BY waiting_since ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
), updated_visitor AS (
    UPDATE visitors
    SET status = 2, updated_at = NOW()
    WHERE public_id = next_visitor_public_id
    RETURNING *
)
SELECT uv.public_id, uv.waiting_since, uv.name, uv.status, uv.daily_ticket_number, uv.purpose_public_id, p.purpose_name
FROM updated_visitor uv
INNER JOIN purposes p
ON p.public_id = uv.purpose_public_id;

-- name: SetAllIncompleteVisitorsStatusAutoCompleted :many
UPDATE visitors
SET STATUS = 6, updated_at = NOW() -- currently autocompleted is status #6. This will remain hardcoded sadly.
WHERE status IN (1, 2, 3) 
RETURNING *;

-- name: SetAllIncompleteVisitorsStatus :many
UPDATE visitors
SET STATUS = $1, updated_at = NOW() -- not sure if I'll need this query honestly
WHERE status IN (1, 2, 3) 
RETURNING *;

