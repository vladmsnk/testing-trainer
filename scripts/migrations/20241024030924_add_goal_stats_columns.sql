-- +goose Up
-- +goose StatementBegin
alter table goals add column is_tracked bool default false;
alter table goal_stats add column created_at timestamp default current_timestamp;
alter table goal_stats add column updated_at timestamp default current_timestamp;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table goal_stats drop column created_at;
alter table goal_stats drop column updated_at;
alter table goals drop column is_tracked;
-- +goose StatementEnd
