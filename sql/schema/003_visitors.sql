-- +goose Up
CREATE TABLE visitors (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    waiting_since TIMESTAMP NOT NULL,
    name TEXT,
    purpose TEXT NOT NULL,
    status INT NOT NULL
);

-- +goose Down
DROP TABLE visitors;