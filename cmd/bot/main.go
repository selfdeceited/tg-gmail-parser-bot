package main

import (
	"context"
	"os"
	"os/signal"

	tgbot "github.com/go-telegram/bot"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram/handlers"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	cfg := loadConfig()

	database, err := db.Connect()
	if err != nil {
		logrus.WithError(err).Fatal("failed to connect to database")
	}

	svc := wireServices(database, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	bot, err := tgbot.New(cfg.TelegramBotToken,
		tgbot.WithDefaultHandler(telegram.ConversationHandler(svc.registrationService, svc.promptService)),
	)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create bot")
	}

	telegram.RegisterHandlers(bot, svc.registrationService, svc.promptService, svc.watchService)

	// Resume watchers for users who had monitoring active before the restart.
	if err := svc.watchService.RestoreAll(ctx, handlers.MakeBotSendFunc(ctx, bot)); err != nil {
		logrus.WithError(err).Error("failed to restore watchers from DB")
	}

	telegram.StartBot(ctx, bot)

	logrus.Info("shutdown: waiting for poller goroutines to exit")
	svc.watchService.Wait()
	logrus.Info("shutdown: complete")
}
