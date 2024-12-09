-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users_time (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    time_offset INT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users_time;
-- +goose StatementEnd
