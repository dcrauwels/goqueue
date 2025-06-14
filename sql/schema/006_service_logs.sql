-- +goose Up
CREATE TABLE service_logs (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    visitor_id UUID NOT NULL,
    user_id UUID NOT NULL,
    desk_id UUID NOT NULL,
    called_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL
);

ALTER TABLE service_logs
    ADD CONSTRAINT fk_visitor FOREIGN KEY (visitor_id) REFERENCES visitors (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_desk FOREIGN KEY (desk_id) REFERENCES desks (id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE service_logs;