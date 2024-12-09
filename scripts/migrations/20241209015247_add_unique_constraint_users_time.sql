-- +goose Up
-- +goose StatementBegin
ALTER TABLE users_time ADD CONSTRAINT unique_username UNIQUE (username);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users_time DROP CONSTRAINT unique_username;
-- +goose StatementEnd
