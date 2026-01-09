-- +goose Up
-- 1. add new columns
ALTER TABLE users
ADD COLUMN desk_public_id TEXT;

-- 2. populate new columns
UPDATE users u
SET desk_public_id = d.public_id
FROM desks d
WHERE u.desk_id = d.id;

-- 3. add new fks
ALTER TABLE users
ADD CONSTRAINT fk_desk_public_id
    FOREIGN KEY (desk_public_id) REFERENCES desks(public_id);

-- 4. clean up old fks and old columns
ALTER TABLE users
DROP CONSTRAINT fk_desk;

ALTER TABLE users
DROP COLUMN desk_id;

-- +goose Down
-- 1. add back old columns
ALTER TABLE users
ADD COLUMN desk_id UUID;

-- 2. populate old columns
UPDATE users up
SET desk_id = d.id
FROM DESKS desk_id
WHERE u.desk_public_id = d.public_id;

-- 3. add back fks for old columns respecting old fk names
ALTER TABLE users
ADD CONSTRAINT fk_desk
    FOREIGN KEY (desk_id) REFERENCES desks(id);

-- 4. cleanup new fks and new columns
ALTER TABLE users
DROP CONSTRAINT fk_desk_public_id;

ALTER TABLE users
DROP COLUMN desk_public_id;