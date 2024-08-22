-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX idx_jobs_source_id_link ON jobs(source_id, link);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_jobs_source_id_link;
-- +goose StatementEnd
