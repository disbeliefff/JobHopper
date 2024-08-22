package fetcher

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/disbeliefff/JobHunter/internal/model"
	"github.com/disbeliefff/JobHunter/internal/source"
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

func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				return err
			}
		}
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, src := range sources {
		wg.Add(1)

		rssSource := source.NewRssSource(src)

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("[ERROR ]fetch %s error: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("[ERROR ]fetch %s error: %v", source.Name(), err)
				return
			}
		}(rssSource)
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []model.Item) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.isItemExcluded(item) {
			continue
		}

		if err := f.jobs.Store(ctx, model.Job{
			SourceID:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (f *Fetcher) isItemExcluded(item model.Item) bool {
	categories := item.Categories

	for _, keyword := range f.filterKeywords {
		keywordLower := strings.ToLower(keyword)
		titleContainsKeyword := strings.Contains(strings.ToLower(item.Title), keywordLower)

		if titleContainsKeyword {
			return true
		}

		for _, category := range categories {
			if strings.Contains(strings.ToLower(category), keywordLower) {
				return true
			}
		}
	}

	return false
}
