-- +goose Up
-- +goose StatementBegin
alter table goals drop column IF EXISTS previous_goal_id;
alter table goals add column IF NOT EXISTS previous_goal_id integer[];
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table goals drop column IF EXISTS previous_goal_id;
alter table goals add column  IF NOT EXISTS previous_goal_id integer;
-- +goose StatementEnd
