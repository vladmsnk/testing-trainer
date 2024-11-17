-- +goose Up
-- +goose StatementBegin
ALTER TABLE goals ADD COLUMN previous_goal_id INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE goals DROP COLUMN previous_goal_id;
-- +goose StatementEnd
