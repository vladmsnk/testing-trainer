-- +goose Up
-- +goose StatementBegin
alter table progress_snapshots add column if not exists start_bound timestamp;
alter table progress_snapshots add column if not exists end_bound timestamp;
update progress_snapshots set start_bound = created_at, end_bound = created_at + interval '1 day' where  start_bound is null;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table progress_snapshots drop column if exists end_bound;
alter table progress_snapshots drop column if exists start_bound;
-- +goose StatementEnd
