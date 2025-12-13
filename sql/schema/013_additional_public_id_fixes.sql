-- +goose Up
ALTER TABLE visitors
DROP COLUMN purpose_id


-- +goose Down
ALTER TABLE visitors
DROP COLUMN
