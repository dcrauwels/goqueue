-- name: GetServiceLogs :many
SELECT * FROM service_logs;

-- name: GetActiveServiceLogs :many
SELECT * FROM service_logs
WHERE is_active = true;

-- name: GetActiveServiceLogsByUserID :one
SELECT * FROM service_logs
where is_active = true AND user_id = $1;