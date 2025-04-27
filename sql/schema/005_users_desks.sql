-- +goose Up
ALTER TABLE users
ADD COLUMN desk_id UUID,
ADD full_name TEXT NOT NULL;

ALTER TABLE users
ADD CONSTRAINT fk_desk FOREIGN KEY (desk_id) REFERENCES desks (id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE users
DROP COLUMN desk_id,
DROP COLUMN full_name;