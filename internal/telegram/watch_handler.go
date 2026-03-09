package telegram

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// WatchHandler handles the /watch command (toggle: start or stop Gmail monitoring).
func WatchHandler(watchSvc service.WatchService) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		if watchSvc.IsWatching(userID) {
			watchSvc.Stop(userID)
			logrus.WithField("user_id", userID).Info("watch: user stopped watching")
			sendText(ctx, b, chatID, "🔴 Gmail monitoring stopped\\.")
			return
		}

		send := MakeBotSendFunc(ctx, b)
		if err := watchSvc.Start(ctx, userID, chatID, send); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("watch: failed to start")
			sendText(ctx, b, chatID, "❌ Failed to start monitoring\\. Please try again\\.")
			return
		}

		logrus.WithField("user_id", userID).Info("watch: user started watching")
		sendText(ctx, b, chatID, "🟢 Gmail monitoring started\\. You'll be notified when matching emails arrive\\.")
	}
}

// MakeBotSendFunc returns a SendFunc that delivers messages via the given bot instance.
// Exported so main.go can pass it to WatchService.RestoreAll on startup.
func MakeBotSendFunc(ctx context.Context, b *tgbot.Bot) service.SendFunc {
	return func(chatID int64, msg string) {
		_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      msg,
			ParseMode: models.ParseModeHTML,
		})
		if err != nil {
			logrus.WithError(err).WithField("chat_id", chatID).Error("watch: failed to send notification")
		}
	}
}
