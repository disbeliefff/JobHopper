package config

import (
	"log"
	"sync"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfighcl"
)

type Config struct {
	TelegramBotToken     string        `hcl:"telegram_bot_token" env:"TELEGRAM_BOT_TOKEN" required:"true"`
	DatabaseDSN          string        `hcl:"database_dsn" env:"DATABASE_URL"`
	FetchInterval        time.Duration `hcl:"fetch_interval" env:"FETCH_INTERVAL" default:"10m"`
	NotificationInterval time.Duration `hcl:"notification_interval" env:"NOTIFICATION_INTERVAL" default:"1m"`
	FilterKeywords       []string      `hcl:"filter_keywords" env:"FILTER_KEYWORDS"`
}

var (
	cfg  Config
	once sync.Once
)

func Get() Config {
	once.Do(func() {
		loader := aconfig.LoaderFor(&cfg, aconfig.Config{
			Files: []string{
				"./config.hcl",
				"./config.local.hcl",
				"$HOME/.config/job-hunter-bot/config.hcl",
			},
			FileDecoders: map[string]aconfig.FileDecoder{
				".hcl": aconfighcl.New(),
			},
		})

		if err := loader.Load(); err != nil {
			log.Fatalf("[ERROR] failed to load config: %v", err) // Завершаем программу при ошибке
		}
		if cfg.TelegramBotToken == "" {
			log.Fatalf("[ERROR] TelegramBotToken is not set or loaded")
		}

		if cfg.DatabaseDSN == "" {
			log.Fatalf("[ERROR] DatabaseDSN is not set or loaded")
		}
	})

	return cfg
}
