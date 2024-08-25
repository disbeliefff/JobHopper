package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/disbeliefff/JobHunter/internal/botkit"
	"github.com/disbeliefff/JobHunter/internal/fetcher"
	"github.com/disbeliefff/JobHunter/internal/model"
	"github.com/disbeliefff/JobHunter/internal/notifier"
	"github.com/disbeliefff/JobHunter/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ViewCmdStart(fetcher *fetcher.Fetcher, jobStorage *storage.JobStorage, usersStorage *storage.UserStorage, botAPI *tgbotapi.BotAPI) botkit.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
		chatID := update.FromChat().ID

		if err := usersStorage.StoreChatID(ctx, chatID); err != nil {
			log.Printf("[ERROR] Failed to store chat ID: %v", err)
			return err
		}

		message := tgbotapi.NewMessage(chatID, "Hello! Starting to parse and fetch new vacancies for you...")

		if _, err := bot.Send(message); err != nil {
			log.Printf("[ERROR] Failed to send initial message: %v", err)
			return err
		}

		log.Println("[INFO] Starting parsing process...")
		vacancies, err := fetcher.Start(ctx)
		if err != nil {
			log.Printf("[ERROR] Error during parsing: %v", err)
			return err
		}

		log.Printf("[INFO] Found %d vacancies during parsing", len(vacancies))

		log.Println("[INFO] Parsing completed. Preparing to send completion message.")

		parseCompleteMsg := tgbotapi.NewMessage(chatID, "Parsing complete. Sending vacancies now...")
		if _, err := bot.Send(parseCompleteMsg); err != nil {
			log.Printf("[ERROR] Failed to send parse complete message: %v", err)
			return err
		}

		time.Sleep(1 * time.Second)

		if len(vacancies) == 0 {
			noJobsMessage := tgbotapi.NewMessage(chatID, "No new vacancies found.")
			if _, err := bot.Send(noJobsMessage); err != nil {
				log.Printf("[ERROR] Failed to send no jobs message: %v", err)
			}
			log.Println("[INFO] No new vacancies found. Ending process.")
			return nil
		}

		log.Println("[INFO] Sending found vacancies...")
		for _, vacancy := range vacancies {
			vacancyMsg := FormatVacancyMessage(vacancy)
			message := tgbotapi.NewMessage(chatID, vacancyMsg)
			if _, err := bot.Send(message); err != nil {
				log.Printf("[ERROR] Failed to send vacancy message: %v", err)
			}
		}

		log.Println("[INFO] All vacancies sent successfully")

		log.Println("[INFO] Starting notifier...")
		notify := notifier.New(jobStorage, botAPI, 24*time.Hour, 24*time.Hour, chatID)
		go func() {
			if err := notify.Start(ctx); err != nil {
				log.Printf("[ERROR] Notifier error: %v", err)
			}
		}()

		return nil
	}
}

func FormatVacancyMessage(job model.Job) string {
	return fmt.Sprintf("Title: %s\nURL: %s\nPublished at: %s",
		job.Title, job.Link, job.PublishedAt.Format(time.RFC1123))
}
