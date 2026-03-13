package main

import (
	"gorm.io/gorm"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/claude"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

type services struct {
	registrationService service.RegistrationService
	promptService       service.PromptService
	watchService        service.WatchService
}

func wireServices(db *gorm.DB, cfg config) services {
	claudeClient := claude.NewClient(cfg.ClaudeAPIKey)
	return services{
		registrationService: service.NewRegistrationService(db, cfg.IOTimeout),
		promptService:       service.NewPromptService(db),
		watchService:        service.NewWatchService(db, claudeClient, cfg.IOTimeout),
	}
}
