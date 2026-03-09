package main

import (
	"context"
	"os"
	"os/signal"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/claude"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	cfg := loadConfig()
	token := cfg.TelegramBotToken
	claudeAPIKey := cfg.ClaudeAPIKey

	database, err := db.Connect()
	if err != nil {
		logrus.WithError(err).Fatal("failed to connect to database")
	}

	regSvc := service.NewRegistrationService(database)
	promptSvc := service.NewPromptService(database)
	claudeClient := claude.NewClient(claudeAPIKey)
	watchSvc := service.NewWatchService(database, claudeClient)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b, err := tgbot.New(token,
		tgbot.WithDefaultHandler(telegram.DefaultHandler(regSvc, promptSvc)),
	)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create bot")
	}

	telegram.RegisterHandlers(b, regSvc, promptSvc, watchSvc)

	// Resume watchers for users who had monitoring active before the restart.
	if err := watchSvc.RestoreAll(ctx, telegram.MakeBotSendFunc(ctx, b)); err != nil {
		logrus.WithError(err).Warn("failed to restore watchers from DB")
	}

	_, err = b.SetMyCommands(ctx, &tgbot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "start", Description: "Start the bot and see the setup guide"},
			{Command: "register", Description: "Link a Gmail account for monitoring"},
			{Command: "clearregistration", Description: "Unlink Gmail account and clear credentials"},
			{Command: "configure", Description: "Manage summarization prompts"},
			{Command: "addprompt", Description: "Add or edit a summarization prompt"},
			{Command: "watch", Description: "Start or stop Gmail monitoring"},
		},
	})
	if err != nil {
		logrus.WithError(err).Warn("failed to set bot commands")
	}

	logrus.Info("bot started, waiting for updates...")
	b.Start(ctx)
}
