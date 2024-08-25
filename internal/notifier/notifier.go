package notifier

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/disbeliefff/JobHunter/internal/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type JobProvider interface {
	AllNotPosted(ctx context.Context, since time.Time) ([]model.Job, error)
	MarkJobPosted(ctx context.Context, id int, chatID int64) error
}

type Notifier struct {
	jobs             JobProvider
	bot              *tgbotapi.BotAPI
	sendInterval     time.Duration
	lookupTimeWindow time.Duration
	chatID           int64 // Добавьте сюда chatID пользователя
}

func New(jobs JobProvider, bot *tgbotapi.BotAPI, sendInterval time.Duration, lookupTimeWindow time.Duration, chatID int64) *Notifier {
	return &Notifier{
		jobs:             jobs,
		bot:              bot,
		sendInterval:     sendInterval,
		lookupTimeWindow: lookupTimeWindow,
		chatID:           chatID, // Установите chatID
	}
}

func (n *Notifier) Start(ctx context.Context) error {
	ticker := time.NewTicker(n.sendInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := n.SendJobs(ctx); err != nil {
				log.Printf("[ERROR] Error sending jobs: %v", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (n *Notifier) SendJobs(ctx context.Context) error {
	jobs, err := n.jobs.AllNotPosted(ctx, time.Now().Add(-n.lookupTimeWindow))
	if err != nil {
		log.Printf("[ERROR] Error fetching jobs: %v", err)
		return err
	}

	if len(jobs) == 0 {
		message := tgbotapi.NewMessage(n.chatID, "No new jobs found today.")
		_, err = n.bot.Send(message)
		if err != nil {
			log.Printf("[ERROR] Error sending notification: %v", err)
			return err
		}
		return nil
	}

	for _, job := range jobs {
		if err := n.sendJob(job); err != nil {
			log.Printf("[ERROR] Error sending job: %v", err)
			return err
		}

		if err := n.jobs.MarkJobPosted(ctx, job.ID, n.chatID); err != nil {
			log.Printf("[ERROR] Error marking job as posted: %v", err)
			return err
		}
	}

	return nil
}

func (n *Notifier) sendJob(job model.Job) error {
	message := tgbotapi.NewMessage(n.chatID, fmt.Sprintf("Job: %s\nLink: %s", job.Title, job.Link))
	_, err := n.bot.Send(message)
	return err
}
