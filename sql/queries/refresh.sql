-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    NULL
)
RETURNING *;

-- name: GetRefreshTokens :many
SELECT * FROM refresh_tokens;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: GetRefreshTokensByUserID :many
SELECT * FROM refresh_tokens
WHERE user_id = $1 AND expires_at > NOW() AND revoked_at IS NULL;

-- name: RevokeRefreshTokenByToken :one
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE TOKEN = $1
RETURNING *;

-- name: RevokeRefreshTokenByUserID :many 
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE user_id = $1 AND expires_at > NOW() AND revoked_at IS NULL -- Also corrected logic (see below)
RETURNING *;