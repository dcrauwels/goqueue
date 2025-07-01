-- +goose Up
CREATE TABLE purposes (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    purpose_name TEXT NOT NULL,
    parent_purpose_id UUID NOT NULL
);

-- add fk constraint to purposes to allow for subpurposes
ALTER TABLE purposes
    ADD CONSTRAINT fk_parent_purpose_id FOREIGN KEY (parent_purpose_id) REFERENCES purposes(id) ON DELETE CASCADE;

-- three step modification to visitors to refit the purpose column to a uuid holder. will write inner join later
ALTER TABLE visitors
    RENAME COLUMN purpose TO purpose_id;

ALTER TABLE visitors
    ALTER COLUMN purpose_id TYPE UUID USING purpose_id::uuid;

ALTER TABLE visitors
    ADD CONSTRAINT fk_purpose FOREIGN KEY (purpose_id) REFERENCES purposes(id);

-- +goose Down
-- Drop foreign key constraint from visitors
ALTER TABLE visitors
    DROP CONSTRAINT fk_purpose;

-- Drop self-referencing constraint from purposes
ALTER TABLE purposes
    DROP CONSTRAINT fk_parent_purpose_id;

-- Drop the purposes table
DROP TABLE purposes;

-- Reverse the visitors table changes (if you want complete rollback)
ALTER TABLE visitors
    ALTER COLUMN purpose_id TYPE TEXT;

ALTER TABLE visitors
    RENAME COLUMN purpose_id TO purpose;