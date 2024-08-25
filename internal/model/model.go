package model

import "time"

// RSS item
type Item struct {
	Title      string
	Categories []string
	Link       string
	Date       time.Time
	Summary    string
	SourceName string
}

type Source struct {
	ID        int
	Name      string
	FeedURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Job struct {
	ID          int
	SourceID    int
	Title       string
	Link        string
	Summary     string
	PublishedAt time.Time
	PostedAt    time.Time
	CreatedAt   time.Time
}

type User struct {
	ID     int64
	ChatID int64
}
