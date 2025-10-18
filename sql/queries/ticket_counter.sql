-- name: update_ticket_counter :one
INSERT INTO ticket_counter (counter_date, last_ticket_number)
VALUES (CURRENT_DATE, 1)
ON CONFLICT (counter_date)
DO UPDATE SET
  last_ticket_number = ticket_counter.last_ticket_number + 1
RETURNING last_ticket_number;