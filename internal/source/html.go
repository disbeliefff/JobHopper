package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

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
	resp, err := http.Get(s.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rssFeed := s.CreateRSSFromHTML(string(body))

	items := lo.Map(rssFeed.Items, func(item *rss.Item, _ int) model.Item {
		return model.Item{
			Title:      item.Title,
			Link:       item.Link,
			Date:       item.Date,
			Summary:    item.Summary,
			SourceName: s.URL,
		}
	})

	return items, nil
}

func (s *HTMLToRSSSource) CreateRSSFromHTML(htmlContent string) *rss.Feed {
	feed := &rss.Feed{
		Title:       "RSS лента из HTML",
		Link:        s.URL,
		Description: "Лента создана из HTML-страницы",
		Categories:  []string{"HTML", "RSS"},
		Items: []*rss.Item{
			{
				Title:   "HTML страница",
				Link:    s.URL,
				Summary: htmlContent,
				Date:    time.Now(),
			},
		},
	}

	return feed
}

func (s *HTMLToRSSSource) ConvertToRSS(ctx context.Context) (*rss.Feed, error) {
	items, err := s.Fetch(ctx)
	if err != nil {
		return nil, err
	}

	rssFeed := s.CreateRSSFromHTML(items[0].Summary)
	return rssFeed, nil
}
