package main

import (
	"context"
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
	"github.com/robfig/cron/v3"
)

func main() {
	// Настройка бота
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create bot: %v", err)
	}
	log.Printf("[INFO] Authorized on account %s", botAPI.Self.UserName)

	// Подключение к базе данных
	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация хранилищ и парсера
	jobStorage := storage.NewJobStorage(db)
	userStorage := storage.NewUserStorage(db)
	sourceStorage := storage.NewSourceStorage(db)
	fetcher := fetcher.New(
		jobStorage,
		sourceStorage,
		config.Get().FetchInterval,
		config.Get().FilterKeywords,
	)

	// Создание контекста для обработки сигналов
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Настройка бота для обработки команд
	jobsBot := botkit.New(botAPI)
	jobsBot.RegisterCmdView("start", bot.ViewCmdStart(fetcher, jobStorage, userStorage, botAPI))

	// Настройка планировщика
	c := cron.New()
	c.AddFunc("0 8,18 * * *", func() {
		log.Println("[INFO] Running scheduled job at 8:00 or 18:00")
		vacancies, err := fetcher.Start(ctx)
		if err != nil {
			log.Printf("[ERROR] Error during scheduled parsing: %v", err)
			return
		}
		log.Printf("[INFO] Found %d vacancies during parsing", len(vacancies))
	})
	c.Start()

	// Запуск бота
	go func() {
		if err := jobsBot.Run(ctx); err != nil {
			log.Fatalf("[ERROR] Bot stopped: %v", err)
		}
	}()

	// Ожидание завершения работы
	<-ctx.Done()
	c.Stop()
	log.Println("[INFO] Shutting down...")
}
