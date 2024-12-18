-- +goose Up
-- +goose StatementBegin
ALTER TABLE progress_snapshots ADD COLUMN IF NOT EXISTS progress_id int not null default 0;
alter table progress_snapshots add column if not exists goal_id int not null default 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE progress_snapshots DROP COLUMN IF EXISTS progress_id;
-- +goose StatementEnd
