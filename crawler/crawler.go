package crawler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Job struct {
	Title   string
	Company string
	JobURL  string
}

type Fetcher interface {
	Fetch(url string) (body string, urls []string, jobs []Job, err error)
}

type RealFetcher struct{}

func (f *RealFetcher) Fetch(url string) (string, []string, []Job, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, nil, fmt.Errorf("error fetching URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, nil, fmt.Errorf("error fetching URL, status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", nil, nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	var jobs []Job
	var urls []string

	doc.Find(".f-vacancy").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".fd-beefy-gunso").Text()
		if strings.Contains(strings.ToLower(title), "golang") {
			company := s.Find(".f-text-gray").Text()
			jobURL, exists := s.Find("a").Attr("href")
			if exists {
				jobs = append(jobs, Job{
					Title:   title,
					Company: company,
					JobURL:  jobURL,
				})
			}
		}
	})

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists && strings.HasPrefix(link, "http") {
			urls = append(urls, link)
		}
	})

	return "", urls, jobs, nil
}

// Crawl рекурсивно обходит страницы начиная с url до максимальной глубины depth
func Crawl(url string, depth int, fetcher Fetcher, jobsChan chan<- Job) {
	if depth <= 0 {
		return
	}
	_, urls, jobs, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, job := range jobs {
		jobsChan <- job
	}

	for _, u := range urls {
		Crawl(u, depth-1, fetcher, jobsChan)
	}
}
