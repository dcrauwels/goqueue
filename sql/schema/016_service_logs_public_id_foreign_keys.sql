-- +goose Up
-- 1. add new columns
ALTER TABLE service_logs
ADD COLUMN user_public_id TEXT,
ADD COLUMN visitor_public_id TEXT,
ADD COLUMN desk_public_id TEXT;

-- 2. populate new columns
UPDATE service_logs s
SET user_public_id = u.public_id
FROM users u
WHERE s.user_id = u.id;

UPDATE service_logs s
SET visitor_public_id = v.public_id
FROM visitors v
WHERE s.visitor_id = v.id;

UPDATE service_logs s
SET desk_public_id = d.public_id
FROM desks d
WHERE s.desk_id = d.id;

-- 3. set new columns not null
ALTER TABLE service_logs
ALTER COLUMN user_public_id SET NOT NULL,
ALTER COLUMN visitor_public_id SET NOT NULL,
ALTER COLUMN desk_public_id SET NOT NULL;

-- 4. add new fks
ALTER TABLE service_logs
ADD CONSTRAINT fk_user_public_id
    FOREIGN KEY (user_public_id) REFERENCES users(public_id),
ADD CONSTRAINT fk_visitor_public_id
    FOREIGN KEY (visitor_public_id) REFERENCES visitors(public_id),
ADD CONSTRAINT fk_desk_public_id
    FOREIGN KEY (desk_public_id) REFERENCES desks(public_id);

-- 5. cleanup old fks and columns
ALTER TABLE service_logs
DROP CONSTRAINT fk_user,
DROP CONSTRAINT fk_visitor,
DROP CONSTRAINT fk_desk;

ALTER TABLE service_logs
DROP COLUMN user_id,
DROP COLUMN visitor_id,
DROP COLUMN desk_id;

-- +goose Down
-- 1. add back old columns
ALTER TABLE service_logs
ADD COLUMN user_id UUID,
ADD COLUMN visitor_id UUID,
ADD COLUMN desk_id UUID;

-- 2. populate old columns
UPDATE service_logs s
SET user_id = u.id
FROM users u
WHERE s.user_public_id = u.public_id;

UPDATE service_logs s
SET visitor_id = v.id
FROM visitors v
WHERE s.visitor_public_id = v.public_id;

UPDATE service_logs s
SET desk_id = d.id
FROM desks d
WHERE s.desk_public_id = d.public_id;

-- 3. set old columns not null
ALTER TABLE service_logs
ALTER COLUMN user_id SET NOT NULL,
ALTER COLUMN visitor_id SET NOT NULL,
ALTER COLUMN desk_id SET NOT NULL;

-- 4. add fks for old columns respecting previous fk names
ALTER TABLE service_logs
ADD CONSTRAINT fk_user
    FOREIGN KEY (user_id) REFERENCES users(id),
ADD CONSTRAINT fk_visitor
    FOREIGN KEY (visitor_id) REFERENCES visitors(id),
ADD CONSTRAINT fk_desk
    FOREIGN KEY (desk_id) REFERENCES desks(pid);

-- 5. cleanup new fks and new columns
ALTER TABLE service_logs
DROP CONSTRAINT fk_user_public_id,
DROP CONSTRAINT fk_visitor_public_id,
DROP CONSTRAINT fk_desk_public_id;

ALTER TABLE service_logs
DROP COLUMN user_public_id,
DROP COLUMN visitor_public_id, 
DROP COLUMN desk_public_id;