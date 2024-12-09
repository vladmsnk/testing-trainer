-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS progress_snapshots (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    progress_ids INT[] NOT NULL, -- идентификаторы прогрессов
    created_at TIMESTAMP NOT NULL -- время создания снимка
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS progress_snapshots;
-- +goose StatementEnd
