package main

import (
	"flag"
	"log"

	"github.com/disbeliefff/JobHunter/internal/telegram"
)

func main() {
	tgClient = telegram.NewClient(mustToken(), tgHost())
}

func mustToken() string {
	token := flag.String("token-bot", "", "token for telegram bot")

	flag.Parse()

	if *token == "" {
		log.Fatal("token is empty")
	}

	return *token
}

func tgHost() string {
	return "api.telegram.org"
}
