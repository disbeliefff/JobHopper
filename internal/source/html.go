package source

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/SlyMarbo/rss"
	"github.com/disbeliefff/JobHunter/internal/model"
	"github.com/samber/lo"
)

type HTMLToRSSSource struct {
	URL string
}

func NewHTMLToRssSource(m model.Source) HTMLToRSSSource {
	return HTMLToRSSSource{
		URL: m.FeedURL,
	}
}

func (s HTMLToRSSSource) Fetch(ctx context.Context) ([]model.Item, error) {
	var allVacancies []model.Item

	nextURL := s.URL

	for nextURL != "" {
		log.Printf("Fetching URL: %s", nextURL)
		vacancies, nextPageURL, err := s.fetchPage(nextURL)
		if err != nil {
			log.Printf("Error fetching URL: %s, error: %v", nextURL, err)
			return nil, err
		}

		allVacancies = append(allVacancies, vacancies...)
		nextURL = nextPageURL
		if nextPageURL != "" {
			log.Printf("Next page URL: %s", nextPageURL)
		}
	}

	return allVacancies, nil
}

func (s *HTMLToRSSSource) fetchPage(url string) ([]model.Item, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching URL: %s, error: %v", url, err)
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("error: status code %d", resp.StatusCode)
		log.Printf("Non-OK HTTP status: %s, error: %v", url, err)
		return nil, "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, "", err
	}

	log.Printf("HTML content received from %s: %s", url, string(body)[:500]) // Показываем первые 500 символов

	vacancies, err := s.extractVacanciesFromHTML(string(body))
	if err != nil {
		log.Printf("Error extracting vacancies: %v", err)
		return nil, "", err
	}

	nextPageURL := s.extractNextPageURL(string(body))

	return vacancies, nextPageURL, nil
}

func (s *HTMLToRSSSource) extractVacanciesFromHTML(htmlContent string) ([]model.Item, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("Error creating document from HTML: %v", err)
		return nil, err
	}

	var vacancies []model.Item

	if strings.Contains(s.URL, "delucru.md") {
		doc.Find(".job-body").Each(func(i int, sel *goquery.Selection) {
			title := sel.Find(".job-title").Text()
			link, exists := sel.Find(".job-title a").Attr("href")
			if !exists {
				log.Printf("Error: Missing link for vacancy #%d", i)
				return
			}

			if !strings.HasPrefix(link, "http") {
				link = fmt.Sprintf("%s%s", s.URL, link)
			}
			summary := sel.Find(".job-body").Text()
			dateStr := sel.Find(".job-date").Text()
			date, err := time.Parse("02-01-2006", dateStr)
			if err != nil {
				log.Printf("Error parsing date: %v, using current time", err)
				date = time.Now()
			}

			vacancies = append(vacancies, model.Item{
				Title:      strings.TrimSpace(title),
				Link:       link,
				Date:       date,
				Summary:    strings.TrimSpace(summary),
				SourceName: s.URL,
			})
		})
	}

	if strings.Contains(s.URL, "joblist.md") {
		doc.Find(".page--ads__items__list__detail__item__header__title__link").Each(func(i int, sel *goquery.Selection) {
			title := sel.Text()
			link, exists := sel.Attr("href")
			if !exists {
				log.Printf("Error: Missing link for vacancy #%d", i)
				return
			}

			if !strings.HasPrefix(link, "http") {
				link = fmt.Sprintf("%s%s", s.URL, link)
			}
			summary := sel.Closest(".page--ads__items__list__detail__item__header__title__link").Find(".job-summary").Text()

			vacancies = append(vacancies, model.Item{
				Title:      strings.TrimSpace(title),
				Link:       link,
				Date:       time.Now(),
				Summary:    strings.TrimSpace(summary),
				SourceName: s.URL,
			})
		})
	}

	if strings.Contains(s.URL, "rabota.md") {
		doc.Find(".vacancy-title").Each(func(i int, sel *goquery.Selection) {
			title := sel.Text()
			link, exists := sel.Attr("href")
			if !exists {
				log.Printf("Error: Missing link for vacancy #%d", i)
				return
			}

			if !strings.HasPrefix(link, "http") {
				baseURL := "https://www.rabota.md"
				link = fmt.Sprintf("%s%s", baseURL, link)
			}
			summary := sel.SiblingsFiltered(".vacancy-summary").Text()

			vacancies = append(vacancies, model.Item{
				Title:      strings.TrimSpace(title),
				Link:       link,
				Date:       time.Now(),
				Summary:    strings.TrimSpace(summary),
				SourceName: s.URL,
			})
		})
	}

	if len(vacancies) == 0 {
		err := fmt.Errorf("no vacancies found in HTML content")
		log.Printf("Error: %v", err)
		return nil, err
	}

	log.Printf("Extracted %d vacancies from HTML content", len(vacancies))
	return vacancies, nil
}

func (s *HTMLToRSSSource) extractNextPageURL(htmlContent string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("Error creating document from HTML: %v", err)
		return ""
	}

	nextPage := doc.Find(".pagination-next a").AttrOr("href", "")
	if nextPage != "" && !strings.HasPrefix(nextPage, "http") {
		nextPage = fmt.Sprintf("%s%s", s.URL, nextPage)
	}

	return nextPage
}

func (s *HTMLToRSSSource) ConvertToRSS(ctx context.Context) (*rss.Feed, error) {
	log.Printf("Converting HTML content from %s to RSS", s.URL)
	items, err := s.Fetch(ctx)
	if err != nil {
		log.Printf("Error during fetch: %v", err)
		return nil, err
	}

	rssFeed := &rss.Feed{
		Title:       "RSS лента вакансий",
		Link:        s.URL,
		Description: "Вакансии, извлеченные из HTML-страницы",
		Categories:  []string{"Вакансии", "RSS"},
		Items: lo.Map(items, func(item model.Item, _ int) *rss.Item {
			return &rss.Item{
				Title:   item.Title,
				Link:    item.Link,
				Summary: item.Summary,
				Date:    item.Date,
			}
		}),
	}

	log.Printf("Successfully converted HTML content from %s to RSS", s.URL)
	return rssFeed, nil
}
