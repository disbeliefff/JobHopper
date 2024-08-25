package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/disbeliefff/JobHunter/internal/bot"
	"github.com/disbeliefff/JobHunter/internal/botkit"
	"github.com/disbeliefff/JobHunter/internal/config"
	"github.com/disbeliefff/JobHunter/internal/fetcher"
	"github.com/disbeliefff/JobHunter/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

func main() {
	log.Printf("[INFO] Starting bot with token %s", config.Get().TelegramBotToken)
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Printf("[ERROR] Failed to create bot: %v", err)
		return
	}
	log.Printf("[INFO] Authorized on account %s", botAPI.Self.UserName)

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		log.Printf("[ERROR] Failed to connect to database: %v", err)
		return
	}
	defer db.Close()

	var (
		jobStorage    = storage.NewJobStorage(db)
		userStorage   = storage.NewUserStorage(db)
		sourceStorage = storage.NewSourceStorage(db)
		fetcher       = fetcher.New(
			jobStorage,
			sourceStorage,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)
	)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	jobsBot := botkit.New(botAPI)

	jobsBot.RegisterCmdView("start", bot.ViewCmdStart(fetcher, jobStorage, userStorage, botAPI))
	log.Println("[INFO] Registered 'start' command")

	// Запуск бота
	go func(ctx context.Context) {
		if err := jobsBot.Run(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR] Error to run bot: %v", err)
				return
			}
			log.Printf("[INFO] Bot stopped")
		}
	}(ctx)

	// Ожидание завершения всех задач
	<-ctx.Done()
	log.Println("[INFO] Application shutting down.")
}
