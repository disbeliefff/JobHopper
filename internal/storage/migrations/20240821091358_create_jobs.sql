-- +goose Up
-- +goose StatementBegin
CREATE TABLE jobs (
     id SERIAL PRIMARY KEY,
     source_id INT NOT NULL,
     title VARCHAR(255) NOT NULL,
     link VARCHAR(255) NOT NULL,
     summary TEXT NOT NULL,
     published_at TIMESTAMP NOT NULL,
     created_at TIMESTAMP NOT NULL,
     posted_at TIMESTAMP,
     
     CONSTRAINT fk_jobs_source_id FOREIGN KEY (source_id) REFERENCES sources(id)

);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
