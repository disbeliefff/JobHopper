-- +goose Up
ALTER TABLE users ADD CONSTRAINT unique_chat_id UNIQUE (chat_id);

-- +goose Down
ALTER TABLE users DROP CONSTRAINT unique_chat_id;
