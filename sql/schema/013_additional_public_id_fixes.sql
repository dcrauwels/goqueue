-- +goose Up
ALTER TABLE visitors
ADD COLUMN purpose_public_id TEXT NOT NULL;

UPDATE visitors v
SET purpose_public_id = p.public_id
FROM purposes p
WHERE v.purpose_id = p.id;

ALTER TABLE visitors
ADD CONSTRAINT fk_public_purpose_id
FOREIGN KEY (purpose_public_id)
REFERENCES purposes(public_id);

ALTER TABLE visitors
DROP CONSTRAINT fk_purpose;

ALTER TABLE visitors
DROP COLUMN purpose_id;

-- +goose Down
ALTER TABLE visitors
ADD COLUMN purpose_id UUID NOT NULL;

UPDATE visitors
SET purpose_id = p.id
FROM purposes p
WHERE v.purpose_public_id = p.public_id

ALTER TABLE visitors
ADD CONSTRAINT fk_purpose 
FOREIGN KEY (purpose_id) 
REFERENCES purposes(id);

ALTER TABLE visitors
DROP CONSTRAINT fk_public_purpose_id;

ALTER TABLE visitors
DROP COLUMN purpose_public_id
