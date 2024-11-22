-- +goose Up
-- +goose StatementBegin
ALTER TABLE goals DROP COLUMN is_tracked;
ALTER TABLE goals DROP COLUMN completed_times;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table goals add column is_tracked bool;
alter table goals add column completed_times integer;
-- +goose StatementEnd
