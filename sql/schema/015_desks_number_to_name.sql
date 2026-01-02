-- +goose Up
ALTER TABLE desks
ADD COLUMN name TEXT;

UPDATE desks d
SET name = number::TEXT;

ALTER TABLE desks
ALTER COLUMN name SET NOT NULL;

ALTER TABLE desks
DROP COLUMN number;

-- +goose Down
ALTER TABLE desks
ADD COLUMN number INT;

UPDATE desks d
SET number = 0;

ALTER TABLE desks
ALTER COLUMN number SET NOT NULL;

ALTER TABLE desks
DROP COLUMN name;