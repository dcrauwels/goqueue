-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password, full_name, is_admin, is_active)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3,
    FALSE,
    TRUE
)
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
where id = $1;

-- name: SetEmailPasswordByID :one
UPDATE users
SET email = $2, hashed_password = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetFullNameByID :one
UPDATE users
SET full_name = $2
where id = $1
RETURNING *;

-- name: SetIsAdminByID :one
UPDATE users
SET is_admin = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetInactiveByID :one
UPDATE users
SET is_active = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUserByID :one
DELETE FROM users
WHERE id = $1
RETURNING *;