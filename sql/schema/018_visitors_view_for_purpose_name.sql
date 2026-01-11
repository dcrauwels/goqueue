-- +goose Up
CREATE VIEW visitor_response_values AS
SELECT
    v.public_id,
    v.waiting_since,
    v.name,
    v.status,
    v.daily_ticket_number,
    v.purpose_public_id,
    p.purpose_name
FROM visitors v
INNER JOIN purposes p
ON p.public_id = v.purpose_public_id;

-- + goose Down
DROP VIEW IF EXISTS visitor_response_values;