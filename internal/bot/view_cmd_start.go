package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
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

		message := tgbotapi.NewMessage(chatID, "Привет! Полный функционал бота все еще в разработке. На данный момент ищу вакансии по запросу golang и backend")

		if _, err := bot.Send(message); err != nil {
			log.Printf("[ERROR] Failed to send initial message: %v", err)
			return err
		}

		bot.Send(tgbotapi.NewMessage(chatID, "Начинаю парсинг..."))

		log.Println("[INFO] Starting parsing process...")
		vacancies, err := fetcher.Start(ctx)
		if err != nil {
			log.Printf("[ERROR] Error during parsing: %v", err)
			return err
		}

		log.Printf("[INFO] Found %d vacancies during parsing", len(vacancies))

		if len(vacancies) == 0 {
			noJobsMessage := tgbotapi.NewMessage(chatID, "Сегодня новых вакансий не нашлось")
			if _, err := bot.Send(noJobsMessage); err != nil {
				log.Printf("[ERROR] Failed to send no jobs message: %v", err)
			}
			log.Println("[INFO] No new vacancies found. Ending process.")
			return nil
		}

		bot.Send(tgbotapi.NewMessage(chatID, "Вакансии найденные по вашему запросу..."))

		log.Println("[INFO] Sending found vacancies...")
		for _, vacancy := range vacancies {
			postedToChatIDs, err := jobStorage.GetPostedToChatIDs(ctx, vacancy.ID)
			if err != nil {
				log.Printf("[ERROR] Failed to check posted_to_chat_ids: %v", err)
				continue
			}

			log.Printf("[DEBUG] Job ID %d has posted_to_chat_ids: %s", vacancy.ID, postedToChatIDs)

			if postedToChatIDs == fmt.Sprintf("%d", chatID) || strings.Contains(postedToChatIDs, fmt.Sprintf(",%d", chatID)) || strings.Contains(postedToChatIDs, fmt.Sprintf("%d,", chatID)) {
				log.Printf("[INFO] Job with link %s has already been posted to chat %d, skipping...", vacancy.Link, chatID)
				continue
			}

			vacancyMsg := FormatVacancyMessage(vacancy)
			message := tgbotapi.NewMessage(chatID, vacancyMsg)
			if _, err := bot.Send(message); err != nil {
				log.Printf("[ERROR] Failed to send vacancy message: %v", err)
				continue
			}

			if err := jobStorage.MarkJobPosted(ctx, vacancy.ID, chatID); err != nil {
				log.Printf("[ERROR] Failed to mark job as posted: %v", err)
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

		bot.Send(tgbotapi.NewMessage(chatID, "Запускаю таймер на 8:00 и 18:00 каждый день"))

		return nil
	}
}

func FormatVacancyMessage(job model.Job) string {
	return fmt.Sprintf("Title: %s\nURL: %s\nPublished at: %s",
		job.Title, job.Link, job.PublishedAt.Format(time.RFC1123))
}
