package telegram

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// ConfigureButtonHandler handles the "⚙️ Configure" reply keyboard button press.
// It re-verifies stored credentials with a live Gmail smoke test:
// - valid → show registration active status
// - invalid or missing → prompt user to run /clearregistration manually
func ConfigureButtonHandler(svc service.RegistrationService) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.From == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		if err := svc.VerifyCredentials(ctx, userID); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Warn("configure: credential re-verification failed")
			sendText(ctx, b, chatID, "❌ Your Gmail credentials could not be verified\\. Please run /clearregistration and then /register again\\.")
			return
		}

		logrus.WithField("user_id", userID).Info("configure: credential re-verification passed")
		sendText(ctx, b, chatID, "✅ Your Gmail account is linked and active!")
	}
}
