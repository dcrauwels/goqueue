-- name: CreatePurpose :one
INSERT INTO purposes (id, public_id, created_at, updated_at, purpose_name, parent_purpose_id)
VALUES (
    gen_random_uuid(),
    $1,
    NOW(),
    NOW(),
    $2,
    $3
)
RETURNING *;

-- name: GetPurposesByPublicID :one
SELECT * FROM purposes
WHERE public_id = $1;

-- name: GetPurposesByName :one
SELECT * FROM purposes
WHERE purpose_name = $1;

-- name: GetPurposes :many
SELECT * FROM purposes;

-- name: GetPurposesByID :one
SELECT * FROM purposes
WHERE id = $1;

-- name: GetPurposesByParent :many
SELECT * FROM purposes
WHERE parent_purpose_id = $1;

-- name: SetPurpose :one
UPDATE purposes
SET purpose_name = $2, parent_purpose_id = $3
WHERE id = $1
RETURNING *;

-- name: SetPurposeName :one
UPDATE purposes
SET purpose_name = $2
WHERE id = $1
RETURNING *;

-- name: SetPurposeParentID :one
UPDATE purposes
SET parent_purpose_id = $2
WHERE id = $1
RETURNING *;

-- name: SetPurposeParentIDByParentPurposeName :one
UPDATE purposes
SET parent_purpose_id = (SELECT purposes.id FROM purposes WHERE purposes.purpose_name = $2)
WHERE purposes.id = $1
RETURNING *;