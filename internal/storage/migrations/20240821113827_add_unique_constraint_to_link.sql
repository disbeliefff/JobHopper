-- +goose Up
-- +goose StatementBegin
ALTER TABLE jobs
ALTER COLUMN link SET NOT NULL;

CREATE UNIQUE INDEX idx_jobs_link ON jobs(link);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_jobs_link;

ALTER TABLE jobs
ALTER COLUMN link DROP NOT NULL;
-- +goose StatementEnd
