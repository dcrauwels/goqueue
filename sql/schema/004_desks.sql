-- +goose Up
CREATE TABLE desks (
    id UUID PRIMARY KEY,
    number INT UNIQUE NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL
);

-- +goose Down
DROP TABLE desks;