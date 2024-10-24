-- +goose Up
-- +goose StatementBegin
alter table goals rename column total_tracking_days to total_tracking_periods;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table goals rename column total_tracking_periods to total_tracking_days;
-- +goose StatementEnd
