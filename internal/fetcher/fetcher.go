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

func (f *Fetcher) Start(ctx context.Context) ([]model.Job, error) {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	// Выполняем первичный парсинг
	jobs, err := f.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Initial fetch returned %d jobs", len(jobs))

	// Возвращаем результат после первичного парсинга и не продолжаем цикл
	return jobs, nil
}

func (f *Fetcher) Fetch(ctx context.Context) ([]model.Job, error) {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var jobs []model.Job

	for _, src := range sources {
		wg.Add(1)

		rssSource := source.NewRssSource(src)

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("[ERROR] fetch %s error: %v", source.Name(), err)
				return
			}

			sourceJobs, err := f.processItems(ctx, source, items)
			if err != nil {
				log.Printf("[ERROR] fetch %s error: %v", source.Name(), err)
				return
			}

			mu.Lock()
			jobs = append(jobs, sourceJobs...)
			mu.Unlock()
		}(rssSource)
	}

	wg.Wait()

	return jobs, nil
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []model.Item) ([]model.Job, error) {
	var jobs []model.Job

	for _, item := range items {
		item.Date = item.Date.UTC()

		if !f.isItemIncluded(item) {
			continue
		}

		job := model.Job{
			SourceID:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
		}

		if err := f.jobs.Store(ctx, job); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (f *Fetcher) isItemIncluded(item model.Item) bool {
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
