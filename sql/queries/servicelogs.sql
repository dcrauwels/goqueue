-- name: GetServiceLogs :many
SELECT * FROM service_logs;

-- name: GetServiceLogsByPublicID :one
SELECT * FROM service_logs
WHERE public_id = $1;

-- name: GetActiveServiceLogs :many
SELECT * FROM service_logs
WHERE is_active = true;

-- name: GetActiveServiceLogsByUserID :many
SELECT * FROM service_logs
where is_active = true AND user_public_id = $1;

-- name: CreateServiceLogs :one
INSERT INTO service_logs (id, public_id, created_at, updated_at, visitor_public_id, user_public_id, desk_public_id, called_at, is_active)
VALUES (
    gen_random_uuid(),
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4,
    NOW(),
    true
)
RETURNING *;

-- name: SetServiceLogsByPublicID :one
UPDATE service_logs
SET visitor_public_id = $2, user_public_id = $3, desk_public_id = $4, is_active = $5, updated_at = NOW()
WHERE public_id = $1
RETURNING *;