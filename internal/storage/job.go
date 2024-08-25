package storage

import (
	"context"
	"fmt"
	"log"
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

	var exists bool
	err = conn.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM jobs WHERE source_id = $1 AND link = $2)`,
		job.SourceID, job.Link)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("[INFO] Job with link %s already exists, skipping insertion.", job.Link)
		return nil
	}

	log.Printf("[INFO] Storing job: %s, %s", job.Title, job.Link)

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

func (j *JobStorage) AllNotPosted(ctx context.Context, since time.Time) ([]model.Job, error) {
	conn, err := j.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var jobs []dbJob
	if err := conn.SelectContext(ctx, &jobs,
		`SELECT * FROM jobs WHERE posted_at IS NULL
        AND published_at >= $1::timestamp
        ORDER BY published_at`, since.UTC().Format(time.RFC3339)); err != nil {
		return nil, err
	}

	log.Printf("[INFO] Found %d jobs that have not been posted", len(jobs))

	return lo.Map(jobs, func(job dbJob, _ int) model.Job {
		return model.Job(job)
	}), nil
}

func (j *JobStorage) MarkJobPosted(ctx context.Context, jobID int, chatID int64) error {
	log.Printf("[INFO] Marking job ID %d as posted for chat ID %d", jobID, chatID)

	_, err := j.db.ExecContext(
		ctx,
		`UPDATE jobs 
         SET posted_to_chat_ids = 
            CASE
                WHEN posted_to_chat_ids IS NULL THEN $1
                WHEN posted_to_chat_ids LIKE '%' || $1 || '%' THEN posted_to_chat_ids
                ELSE posted_to_chat_ids || ',' || $1
            END
         WHERE id = $2`,
		fmt.Sprintf("%d", chatID), jobID,
	)
	return err
}

func (j *JobStorage) GetPostedToChatIDs(ctx context.Context, jobID int) (string, error) {
	var postedToChatIDs string
	err := j.db.GetContext(ctx, &postedToChatIDs, `SELECT posted_to_chat_ids FROM jobs WHERE id = $1`, jobID)
	return postedToChatIDs, err
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
