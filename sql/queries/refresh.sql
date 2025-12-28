-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, public_id, created_at, updated_at, user_public_id, expires_at, revoked_at)
VALUES (
    $1,
    $2,
    NOW(),
    NOW(),
    $3,
    $4,
    NULL
)
RETURNING *;

-- name: GetRefreshTokensByPublicID :one
SELECT * FROM refresh_tokens
WHERE public_id = $1;

-- name: GetRefreshTokens :many
SELECT * FROM refresh_tokens;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: GetRefreshTokensByUserPublicID :many
SELECT * FROM refresh_tokens
WHERE user_public_id = $1 AND expires_at > NOW() AND revoked_at IS NULL;

-- name: RevokeRefreshTokenByToken :one
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = $1
RETURNING *;

-- name: RevokeRefreshTokenByUserPublicID :many 
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE user_public_id = $1 AND expires_at > NOW() AND revoked_at IS NULL
RETURNING *;

-- name: RevokeRefreshTokens :many
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE expires_at > NOW() AND revoked_at IS NULL
returning *;