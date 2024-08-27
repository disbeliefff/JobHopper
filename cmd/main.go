package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/disbeliefff/JobHunter/internal/bot"
	"github.com/disbeliefff/JobHunter/internal/botkit"
	"github.com/disbeliefff/JobHunter/internal/config"
	"github.com/disbeliefff/JobHunter/internal/fetcher"
	"github.com/disbeliefff/JobHunter/internal/notifier"
	"github.com/disbeliefff/JobHunter/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	serverReady := make(chan bool)
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		log.Printf("[INFO] Starting HTTP server on port %s", port)
		serverReady <- true
		if err := http.ListenAndServe(":"+port, nil); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] Failed to start HTTP server: %v", err)
		}
	}()

	<-serverReady

	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Fatalf("[ERROR] Failed to create bot: %v", err)
	}
	log.Printf("[INFO] Authorized on account %s", botAPI.Self.UserName)

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		log.Fatalf("[ERROR] Failed to connect to database: %v", err)
	}
	defer db.Close()

	jobStorage := storage.NewJobStorage(db)
	userStorage := storage.NewUserStorage(db)
	sourceStorage := storage.NewSourceStorage(db)
	fetcher := fetcher.New(
		jobStorage,
		sourceStorage,
		config.Get().FetchInterval,
		config.Get().FilterKeywords,
	)

	jobsBot := botkit.New(botAPI)
	jobsBot.RegisterCmdView("start", func(ctx context.Context, tgbot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		chatID := update.FromChat().ID

		tgbot.Send(tgbotapi.NewMessage(chatID, "Привет! Полный функционал бота все еще в разработке. На данный момент ищу вакансии по запросу golang и backend"))

		if err := userStorage.StoreChatID(ctx, chatID); err != nil {
			log.Printf("[ERROR] Failed to store chat ID: %v", err)
			return err
		}

		tgbot.Send(tgbotapi.NewMessage(chatID, "Начинаю парсинг..."))

		go func() {
			// Создание нового контекста для фетчера
			fetcherCtx, fetcherCancel := context.WithCancel(context.Background())
			defer fetcherCancel()

			vacancies, err := fetcher.Start(fetcherCtx)
			if err != nil {
				log.Printf("[ERROR] Error during initial parsing: %v", err)
				return
			}
			log.Printf("[INFO] Found %d vacancies during initial parsing", len(vacancies))

			if len(vacancies) == 0 {
				tgbot.Send(tgbotapi.NewMessage(chatID, "Сегодня новых вакансий не нашлось"))
			} else {
				tgbot.Send(tgbotapi.NewMessage(chatID, "Вакансии найденные по вашему запросу..."))
				for _, vacancy := range vacancies {
					vacancyMsg := bot.FormatVacancyMessage(vacancy)
					tgbot.Send(tgbotapi.NewMessage(chatID, vacancyMsg))
				}
			}
		}()

		tgbot.Send(tgbotapi.NewMessage(chatID, "Запускаю таймер на 8:00 и 18:00 каждый день"))

		// Запуск периодического фетчера с использованием cron
		go func() {
			c := cron.New()
			c.AddFunc("0 8,18 * * *", func() {
				log.Println("[INFO] Running scheduled job at 8:00 or 18:00")

				// Новый контекст для планового фетчера
				scheduledCtx, scheduledCancel := context.WithTimeout(context.Background(), 30*time.Minute)
				defer scheduledCancel()

				vacancies, err := fetcher.Start(scheduledCtx)
				if err != nil {
					log.Printf("[ERROR] Error during scheduled parsing: %v", err)
					return
				}
				log.Printf("[INFO] Found %d vacancies during scheduled parsing", len(vacancies))

				if len(vacancies) == 0 {
					tgbot.Send(tgbotapi.NewMessage(chatID, "Сегодня новых вакансий не нашлось"))
				} else {
					for _, vacancy := range vacancies {
						vacancyMsg := bot.FormatVacancyMessage(vacancy)
						tgbot.Send(tgbotapi.NewMessage(chatID, vacancyMsg))
					}
				}
			})
			c.Start()
		}()

		go func() {
			log.Println("[INFO] Starting notifier...")
			// Создание нового контекста для нотифаера
			notifyCtx, notifyCancel := context.WithCancel(context.Background())
			defer notifyCancel()

			notify := notifier.New(jobStorage, botAPI, 24*time.Hour, 24*time.Hour, chatID)
			if err := notify.Start(notifyCtx); err != nil {
				log.Printf("[ERROR] Notifier error: %v", err)
			}
		}()

		return nil
	})

	go func() {
		if err := jobsBot.Run(ctx); err != nil {
			log.Fatalf("[ERROR] Bot stopped: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("[INFO] Shutting down...")
}
