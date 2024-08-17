package telegram

import (
	"github.com/disbeliefff/JobHunter/internal/telegram"
)

type Processor struct {
	tg     *telegram.Client
	offset int
	// storage
}

func NewProcessor(tg *telegram.Client) *Processor {
	return &Processor{
		tg:     tg,
		offset: 0,
	}
}
