-- +goose Up
ALTER TABLE refresh_tokens
ADD COLUMN user_public_id TEXT NOT NULL;

UPDATE refresh_tokens t
SET user_public_id = u.public_id
FROM users u
WHERE t.user_id = u.id;

ALTER TABLE refresh_tokens
ADD CONSTRAINT fk_user_public_id
FOREIGN KEY (user_public_id)
REFERENCES users(public_id);

ALTER TABLE refresh_tokens
DROP CONSTRAINT refresh_tokens_user_id_fkey;

ALTER TABLE refresh_tokens
DROP COLUMN user_id;

-- +goose Down
ALTER TABLE refresh_tokens
ADD COLUMN user_id UUID NOT NULL;

UPDATE refresh_tokens t
SET user_id = u.id
FROM users u
WHERE t.user_public_id = u.public_id

ALTER TABLE refresh_tokens
ADD CONSTRAINT refresh_tokens_user_id_fkey 
FOREIGN KEY (user_id) 
REFERENCES users(id);

ALTER TABLE refresh_tokens
DROP CONSTRAINT fk_user_public_id;

ALTER TABLE refresh_tokens
DROP COLUMN user_public_id;
