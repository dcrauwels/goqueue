-- +goose Up
-- users
ALTER TABLE users
ADD COLUMN public_id TEXT UNIQUE NOT NULL;
CREATE INDEX idx_users_public_id ON users(public_id);

-- refresh
ALTER TABLE refresh_tokens
ADD COLUMN public_id TEXT UNIQUE NOT NULL;
CREATE INDEX idx_refresh_tokens_public_id ON refresh_tokens(public_id);

-- purposes
ALTER TABLE purposes
ADD COLUMN public_id TEXT UNIQUE NOT NULL;
CREATE INDEX idx_purposes_public_id ON purposes(public_id);

-- visitors
ALTER TABLE visitors
ADD COLUMN public_id TEXT UNIQUE NOT NULL;
CREATE INDEX idx_visitors_public_id ON visitors(public_id);

-- service_logs
ALTER TABLE service_logs
ADD COLUMN public_id TEXT UNIQUE NOT NULL;
CREATE INDEX idx_service_logs_public_id ON service_logs(public_id);

-- desks
ALTER TABLE desks
ADD COLUMN public_id TEXT UNIQUE NOT NULL;
CREATE INDEX idx_desks_public_id ON desks(public_id);

-- +goose Down
ALTER TABLE users
DROP COLUMN public_id;
DROP INDEX idx_users_public_id;

ALTER TABLE refresh_tokens
DROP COLUMN public_id;
DROP INDEX idx_refresh_tokens_public_id;

ALTER TABLE purposes
DROP COLUMN public_id;
DROP INDEX idx_purposes_public_id;

ALTER TABLE visitors
DROP COLUMN public_id;
DROP INDEX idx_visitors_public_id;

ALTER TABLE service_logs
DROP COLUMN public_id;
DROP INDEX idx_service_logs_public_id;

ALTER TABLE desks
DROP COLUMN public_id;
DROP INDEX idx_desks_public_id;