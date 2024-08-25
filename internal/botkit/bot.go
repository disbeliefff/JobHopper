package botkit

import (
	"context"
	"log"
	"runtime/debug"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	cmdViews map[string]ViewFunc
}

// TODO:add filterKeywords
// add deleteFilterKeywords
// add listOfFilterKeywords

type ViewFunc func(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) error

func New(api *tgbotapi.BotAPI) *Bot {
	return &Bot{
		api: api,
	}
}

func (b *Bot) RegisterCmdView(cmd string, view ViewFunc) {
	if b.cmdViews == nil {
		b.cmdViews = make(map[string]ViewFunc)
	}

	b.cmdViews[cmd] = view
}

func (b *Bot) Run(ctx context.Context) error {
	upd := tgbotapi.NewUpdate(0)
	upd.Timeout = 60

	updates := b.api.GetUpdatesChan(upd)

	for {
		select {
		case update := <-updates:
			log.Printf("[INFO] Received update: %+v", update)
			updateCtx, updateCancel := context.WithCancel(ctx)
			b.handleUpdate(updateCtx, &update)
			updateCancel()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update *tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("[ERROR] Panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	if update.Message == nil {
		log.Println("[INFO] No message in update")
		return
	}

	if !update.Message.IsCommand() {
		log.Println("[INFO] Not a command message")
		return
	}

	cmd := update.Message.Command()
	log.Printf("[INFO] Received command: %s", cmd)

	view, ok := b.cmdViews[cmd]
	if !ok {
		log.Printf("[INFO] Command not registered: %s", cmd)
		return
	}

	if err := view(ctx, b.api, update); err != nil {
		log.Printf("[ERROR] Failed to handle update: %v", err)

		if _, err := b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "internal error")); err != nil {
			log.Printf("[ERROR] Failed to send message: %v", err)
		}
	}
}
