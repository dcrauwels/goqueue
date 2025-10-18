-- +goose Up
CREATE TABLE ticket_counter (
    date TIMESTAMP PRIMARY KEY,
    last_ticket_number INTEGER NOT NULL
);

-- +goose Down
DROP TABLE ticket_counter;