-- +goose Up
-- +goose StatementBegin
alter table goals add column next_check_date timestamp default current_timestamp;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table goals drop column next_check_date;
-- +goose StatementEnd
