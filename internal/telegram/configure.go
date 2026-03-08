package telegram

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/commands"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/queries"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/gmail"
)

// ConfigureButtonHandler handles the "⚙️ Configure" reply keyboard button press.
// It re-verifies stored credentials with a live Gmail smoke test:
// - valid → show registration active status
// - invalid or missing → clear registration, prompt re-register
func ConfigureButtonHandler(db *gorm.DB) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.From == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		creds, err := queries.GetCredentials(db, userID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Warn("configure: no credentials found — clearing registration")
			clearRegistration(ctx, b, db, userID, chatID)
			return
		}

		if err := gmail.VerifyRefreshToken(ctx, creds.ClientID, creds.ClientSecret, creds.RefreshToken); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Warn("configure: credential re-verification failed — clearing")
			clearRegistration(ctx, b, db, userID, chatID)
			return
		}

		logrus.WithField("user_id", userID).Info("configure: credential re-verification passed")
		sendText(ctx, b, chatID,
			"✅ Your Gmail account is linked and active\\.\n\n"+
				"Use /configure to manage your email monitoring settings\\.\n"+
				"To unlink your account, use /clearregistration\\.",
		)
	}
}

// clearRegistration deletes stored credentials, unmarks registration, and tells the user to re-register.
func clearRegistration(ctx context.Context, b *tgbot.Bot, db *gorm.DB, userID, chatID int64) {
	if err := commands.DeleteCredential(db, userID); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("configure: failed to delete credential")
	}
	if err := commands.SetRegistered(db, userID, false); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("configure: failed to unmark registration")
	}
	logrus.WithField("user_id", userID).Info("configure: registration cleared")

	_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    chatID,
		Text:      "❌ Your Gmail credentials are no longer valid\\. Your account has been unlinked\\.\n\nPlease run /register to reconnect\\.",
		ParseMode: models.ParseModeMarkdown,
		ReplyMarkup: &models.ReplyKeyboardRemove{
			RemoveKeyboard: true,
		},
	})
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("configure: failed to send cleared message")
	}
}
