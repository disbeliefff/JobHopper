package crawler

import (
	"fmt"
	"testing"
)

// MockFetcher для тестирования
type MockFetcher struct {
	pages map[string]*fakeResult
}

type fakeResult struct {
	body string
	urls []string
	jobs []Job
}

func (f *MockFetcher) Fetch(url string) (string, []string, []Job, error) {
	if res, ok := f.pages[url]; ok {
		return res.body, res.urls, res.jobs, nil
	}
	return "", nil, nil, fmt.Errorf("not found: %s", url)
}

func TestCrawl(t *testing.T) {
	mockFetcher := &MockFetcher{
		pages: map[string]*fakeResult{
			"https://test.com": {
				body: "",
				urls: []string{"https://test.com/page1"},
				jobs: []Job{
					{Title: "Golang Developer", Company: "TestCorp", JobURL: "https://test.com/job1"},
				},
			},
			"https://test.com/page1": {
				body: "",
				urls: []string{},
				jobs: []Job{
					{Title: "Senior Golang Developer", Company: "TestCorp", JobURL: "https://test.com/job2"},
				},
			},
		},
	}

	jobsChan := make(chan Job)
	go func() {
		defer close(jobsChan)
		Crawl("https://test.com", 2, mockFetcher, jobsChan)
	}()

	var jobs []Job
	for job := range jobsChan {
		jobs = append(jobs, job)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	expectedJobs := []Job{
		{Title: "Golang Developer", Company: "TestCorp", JobURL: "https://test.com/job1"},
		{Title: "Senior Golang Developer", Company: "TestCorp", JobURL: "https://test.com/job2"},
	}

	for i, job := range jobs {
		if job != expectedJobs[i] {
			t.Errorf("expected job %v, got %v", expectedJobs[i], job)
		}
	}
}
