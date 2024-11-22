-- +goose Up
-- +goose StatementBegin
alter table goals drop column previous_goal_id;
alter table goals add column previous_goal_id integer[];
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table goals drop column previous_goal_id;
alter table goals add column previous_goal_id integer;
-- +goose StatementEnd
