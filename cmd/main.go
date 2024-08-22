package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/disbeliefff/JobHunter/internal/config"
	"github.com/disbeliefff/JobHunter/internal/fetcher"
	"github.com/disbeliefff/JobHunter/internal/notifier"
	"github.com/disbeliefff/JobHunter/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

func main() {
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Printf("[ERROR]error to create bot: %v", err)
		return
	}

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		log.Printf("[ERROR]error to connect to database: %v", err)
		return
	}
	defer db.Close()

	var (
		jobStorage    = storage.NewJobStorage(db)
		sourceStorage = storage.NewSourceStorage(db)
		fetcher       = fetcher.New(
			jobStorage,
			sourceStorage,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)

		notifier = notifier.New(
			jobStorage,
			botAPI,
			config.Get().NotificationInterval,
			2*config.Get().FetchInterval,
		)
	)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	go func(ctx context.Context) {
		if err := fetcher.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR]error to fetch jobs: %v", err)
				return
			}
			log.Printf("[INFO]fetcher stopped")
		}
	}(ctx)

	if err := notifier.Start(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("[ERROR]error to send jobs: %v", err)
			return
		}
		log.Printf("[INFO]notifier stopped")
	}

}
