-- +goose Up
CREATE TYPE notification_status AS ENUM ('Pending', 'Sent', 'Fail');

ALTER TABLE notifications
    ADD COLUMN status     notification_status NOT NULL DEFAULT 'Pending',
    ADD COLUMN updated_at TIMESTAMPTZ         NOT NULL DEFAULT now();

-- +goose Down
ALTER TABLE notifications
    DROP COLUMN updated_at,
    DROP COLUMN status;

DROP TYPE notification_status;
