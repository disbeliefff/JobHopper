-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX idx_sources_feed_url ON sources(feed_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_sources_feed_url;
-- +goose StatementEnd
