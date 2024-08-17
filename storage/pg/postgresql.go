package pg

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	db, err := sql.Open("postgres", path)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) SaveVacancy(p *Page, v *Vacancy) error {
	_, err := s.db.ExecContext(context.Background(), `
        INSERT INTO vacancies (page_url, title, company, location, date, url)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (url) DO NOTHING
    `, p.URL, v.Title, v.Company, v.Location, v.Date, v.URL)
	return err
}

func (s *Storage) VacancyExists(vacancyURL string) bool {
	var exists bool
	err := s.db.QueryRowContext(context.Background(), `
        SELECT EXISTS (
            SELECT 1 FROM vacancies WHERE url = $1
        )
    `, vacancyURL).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
