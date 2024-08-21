package fetcher

import (
	"context"
	"time"

	"github.com/disbeliefff/JobHunter/internal/model"
)

type JobStorage interface {
	Store(ctx context.Context, job model.Job) error
}

type SourceProvider interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

type Source interface {
	ID() int
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

type Fetcher struct {
	jobs           JobStorage
	sources        SourceProvider
	fetchInterval  time.Duration
	filterKeywords []string
}

func New(jobs JobStorage, sources SourceProvider, fetchInterval time.Duration, filterKeywords []string) *Fetcher {
	return &Fetcher{
		jobs:           jobs,
		sources:        sources,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
	}
}
