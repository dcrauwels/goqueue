-- +goose Up
ALTER TABLE purposes
ALTER COLUMN parent_purpose_id DROP NOT NULL;

-- +goose Down
ALTER TABLE purposes
ALTER COLUMN parent_purpose_id SET NOT NULL;