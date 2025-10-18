-- +goose Up
ALTER TABLE visitors
ADD COLUMN daily_ticket_number INTEGER NOT NULL;

-- +goose Down
ALTER TABLE visitors
DROP COLUMN daily_ticket_number;