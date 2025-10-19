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
where is_active = true AND user_id = $1;

-- name: CreateServiceLogs :one
INSERT INTO service_logs (id, public_id, created_at, updated_at, visitor_id, user_id, desk_id, called_at, is_active)
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