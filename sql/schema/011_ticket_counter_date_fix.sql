-- +goose Up
ALTER TABLE ticket_counter
ALTER COLUMN date TYPE DATE;

-- +goose Down
ALTER TABLE ticket_counter
ALTER COLUMN date TYPE TIMESTAMP;