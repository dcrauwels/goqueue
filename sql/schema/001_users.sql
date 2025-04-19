-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    email TEXT UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL,
    is_admin BOOLEAN NOT NULL,
    is_active BOOLEAN NOT NULL
);

-- +goose Down
DROP TABLE users;