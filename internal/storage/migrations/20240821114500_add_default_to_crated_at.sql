-- +goose Up
-- +goose StatementBegin
ALTER TABLE jobs
ALTER COLUMN created_at SET DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE jobs
ALTER COLUMN created_at DROP DEFAULT;
-- +goose StatementEnd
