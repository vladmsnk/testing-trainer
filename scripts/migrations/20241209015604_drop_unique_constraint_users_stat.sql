-- +goose Up
-- +goose StatementBegin
alter table goal_stats drop constraint goal_stats_goal_id_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table goal_stats add constraint goal_stats_goal_id_key unique (goal_id);
-- +goose StatementEnd
