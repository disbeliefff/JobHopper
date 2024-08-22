package notifier

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/disbeliefff/JobHunter/internal/model"
)

type JobProvider interface {
	AllNotPosted(ctx context.Context, since time.Time) ([]model.Job, error)
	MarkJobPosted(ctx context.Context, id int) error
}

type Notifier struct {
	jobs             JobProvider
	bot              *tgbotapi.BotAPI
	sendInterval     time.Duration
	lookupTimeWindow time.Duration
}

func New(jobs JobProvider, bot *tgbotapi.BotAPI, sendInterval time.Duration, lookupTimeWindow time.Duration) *Notifier {
	return &Notifier{
		jobs:             jobs,
		bot:              bot,
		sendInterval:     sendInterval,
		lookupTimeWindow: lookupTimeWindow,
	}
}

func (n *Notifier) Start(ctx context.Context) error {
	ticker := time.NewTicker(n.sendInterval)
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
		message := tgbotapi.NewMessage(n.bot.Self.ID, "No new jobs found today.")
		_, err = n.bot.Send(message)
		if err != nil {
			log.Printf("[ERROR] Error sending notification: %v", err)
			return err
		}
		return nil
	}

	job := jobs[0]
	if err := n.sendJob(job); err != nil {
		log.Printf("[ERROR] Error sending job: %v", err)
		return err
	}

	if err := n.jobs.MarkJobPosted(ctx, job.ID); err != nil {
		log.Printf("[ERROR] Error marking job as posted: %v", err)
		return err
	}

	return nil
}

func (n *Notifier) sendJob(job model.Job) error {
	message := tgbotapi.NewMessage(n.bot.Self.ID, fmt.Sprintf("Job: %s\nLink: %s", job.Title, job.Link))
	_, err := n.bot.Send(message)
	return err
}
