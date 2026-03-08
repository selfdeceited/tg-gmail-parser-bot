package telegram

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// ClearRegistrationHandler handles /clearregistration — wipes stored credentials and re-shows the setup guide.
func ClearRegistrationHandler(svc service.RegistrationService) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		logrus.WithField("user_id", userID).Info("ClearRegistrationHandler: start")

		if err := svc.ClearCredentials(ctx, userID); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("ClearRegistrationHandler: failed to clear credentials")
			sendText(ctx, b, chatID, "Failed to clear registration data\\. Please try again later\\.")
			return
		}

		logrus.WithField("user_id", userID).Info("ClearRegistrationHandler: credentials cleared")
		sendText(ctx, b, chatID, "Registration cleared\\. Your Gmail credentials have been deleted\\.")

		// Re-show the setup guide so the user knows how to register again.
		_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      startMessage,
			ParseMode: models.ParseModeMarkdown,
		})
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("ClearRegistrationHandler: failed to send start message")
		}
	}
}
