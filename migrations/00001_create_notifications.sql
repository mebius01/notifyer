-- +goose Up
CREATE TABLE notifications (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    recipient  TEXT        NOT NULL,
    subject    TEXT        NOT NULL DEFAULT '',
    message    TEXT        NOT NULL DEFAULT '',
    channel    TEXT        NOT NULL,
    is_urgent  BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE notifications;
