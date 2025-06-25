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