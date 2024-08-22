package storage

import (
	"context"
	"time"

	"github.com/disbeliefff/JobHunter/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
)

type JobStorage struct {
	db *sqlx.DB
}

func NewJobStorage(db *sqlx.DB) *JobStorage {
	return &JobStorage{
		db: db,
	}
}

func (j *JobStorage) Store(ctx context.Context, job model.Job) error {
	conn, err := j.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err = conn.ExecContext(
		ctx,
		`INSERT INTO jobs (source_id, title, link, summary, published_at, posted_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT(source_id, link) DO NOTHING`,
		job.SourceID, job.Title, job.Link, job.Summary, job.PublishedAt, job.PostedAt, job.CreatedAt,
	); err != nil {
		return err
	}

	return nil
}

func (j *JobStorage) AllNotPosted(ctx context.Context, since time.Time, limit uint) ([]model.Job, error) {
	conn, err := j.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var jobs []dbJob
	if err := conn.SelectContext(ctx, &jobs,
		`SELECT * FROM jobs WHERE posted_at IS NULL
	  AND published_at >= $1::timestamp
	  ORDER BY published_at DESC LIMIT $2`, since.UTC().Format(time.RFC3339), limit); err != nil {
		return nil, err
	}

	return lo.Map(jobs, func(job dbJob, _ int) model.Job {
		return model.Job(job)
	}), nil
}

func (j *JobStorage) MarkJobPosted(ctx context.Context, id int) error {
	_, err := j.db.ExecContext(
		ctx,
		`UPDATE jobs
			SET posted_at = $1::timestamp
			WHERE id = $2`,
		time.Now().UTC().Format(time.RFC3339), id,
	)
	return err
}

type dbJob struct {
	ID          int       `db:"id"`
	SourceID    int       `db:"source_id"`
	Title       string    `db:"title"`
	Link        string    `db:"link"`
	Summary     string    `db:"summary"`
	PublishedAt time.Time `db:"published_at"`
	PostedAt    time.Time `db:"posted_at"`
	CreatedAt   time.Time `db:"created_at"`
}
