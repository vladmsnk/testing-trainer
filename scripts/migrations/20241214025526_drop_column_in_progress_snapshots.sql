-- +goose Up
-- +goose StatementBegin
ALTER TABLE progress_snapshots DROP COLUMN IF EXISTS progress_ids;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table progress_snapshots add column if not exists progress_ids INT[] not null default '{}';
-- +goose StatementEnd
