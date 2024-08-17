package storage

import "time"

type Page struct {
	URL      string
	UserName string
}
type Storage interface {
	SaveVacancy(p *Page, v *Vacancy) error         // Сохранить новую вакансию, привязанную к странице
	LoadVacancies(p *Page) ([]Vacancy, error)      // Загрузить все вакансии с конкретной страницы
	VacancyExists(p *Page, vacancyID string) bool  // Проверить, существует ли вакансия на данной странице
	DeleteVacancy(p *Page, vacancyID string) error // Удалить вакансию с определенной страницы
	UpdateVacancy(p *Page, v *Vacancy) error       // Обновить вакансию, привязанную к странице
	GetLatestVacancy(p *Page) (*Vacancy, error)    // Получить последнюю вакансию с определенной страницы
}

type Vacancy struct {
	ID       string    // Идентификатор вакансии
	Title    string    // Название вакансии
	Company  string    // Название компании
	Location string    // Местоположение вакансии
	Date     time.Time // Дата публикации вакансии
	URL      string    // URL вакансии
}
