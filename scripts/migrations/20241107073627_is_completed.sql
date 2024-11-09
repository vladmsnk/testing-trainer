-- +goose Up
-- +goose StatementBegin
alter table goals add column is_completed bool default false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table goals drop column is_completed;
-- +goose StatementEnd
