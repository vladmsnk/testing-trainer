-- +goose Up
-- +goose StatementBegin
ALTER TABLE goals ADD COLUMN IF NOT EXISTS start_tracking_at timestamp default current_timestamp;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE goals DROP COLUMN IF EXISTS start_tracking_at;
-- +goose StatementEnd
