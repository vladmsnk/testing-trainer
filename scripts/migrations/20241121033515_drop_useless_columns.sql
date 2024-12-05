-- +goose Up
-- +goose StatementBegin
ALTER TABLE goals DROP COLUMN IF EXISTS is_tracked;
ALTER TABLE goals DROP COLUMN IF EXISTS completed_times;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table goals add column IF NOT EXISTS is_tracked bool;
alter table goals add column IF NOT EXISTS completed_times integer;
-- +goose StatementEnd
