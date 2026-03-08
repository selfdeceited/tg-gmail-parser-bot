package telegram

import (
	"context"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/commands"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/queries"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/gmail"
)

// RegisterHandler handles the /register command (step 1: ask for credentials.json).
func RegisterHandler(db *gorm.DB) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {

		if update.Message == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		if creds, err := queries.GetCredentials(db, userID); err == nil {
			if err := gmail.VerifyRefreshToken(ctx, creds.ClientID, creds.ClientSecret, creds.RefreshToken); err != nil {
				logrus.WithError(err).WithField("user_id", userID).Warn("register: existing credentials invalid — clearing, allowing re-registration")
				clearRegistration(ctx, b, db, userID, chatID)
				return
			}
			logrus.WithField("user_id", userID).Info("register: already registered with valid credentials, skipping")
			_, _ = b.SendMessage(ctx, &tgbot.SendMessageParams{
				ChatID:    chatID,
				Text:      "✅ You already have a Gmail account linked\\.\n\nUse /configure to manage settings, or /clearregistration to unlink and start over\\.",
				ParseMode: models.ParseModeMarkdown,
				ReplyMarkup: &models.ReplyKeyboardMarkup{
					Keyboard:        [][]models.KeyboardButton{{{Text: "⚙️ Configure"}}},
					ResizeKeyboard:  true,
					OneTimeKeyboard: false,
				},
			})
			return
		}

		setState(userID, &registerState{step: stepWaitCredentials})
		sendText(ctx, b, chatID, registerGuideMessage)
	}
}

// HandleConversation routes non-command messages to active conversation flows.
func HandleConversation(db *gorm.DB) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.From == nil {
			return
		}
		userID := update.Message.From.ID
		s := getState(userID)
		if s == nil {
			return
		}

		switch s.step {
		case stepWaitCredentials:
			handleCredentialsPaste(ctx, b, update, userID, s)
		case stepWaitAuthCode:
			handleAuthCodePaste(ctx, b, update, db, userID, s)
		}
	}
}

func handleCredentialsPaste(ctx context.Context, b *tgbot.Bot, update *models.Update, userID int64, s *registerState) {
	chatID := update.Message.Chat.ID
	raw := strings.TrimSpace(update.Message.Text)

	cfg, err := gmail.ParseCredentials([]byte(raw))
	if err != nil {
		sendText(ctx, b, chatID, "❌ Could not parse credentials\\.json: "+escapeMarkdown(err.Error())+"\n\nPlease paste the full contents of your credentials\\.json file\\.")
		return
	}

	authURL := gmail.BuildAuthURL(cfg)
	s.oauthConfig = cfg
	s.step = stepWaitAuthCode
	setState(userID, s)

	sendText(ctx, b, chatID,
		"✅ Credentials parsed\\.\n\n"+
			"*Step 2 — Authorize Gmail access*\n\n"+
			"Visit the link below, sign in with your Google account, and paste the authorization code here:\n\n"+
			"`"+escapeMarkdown(authURL)+"`",
	)
}

func handleAuthCodePaste(ctx context.Context, b *tgbot.Bot, update *models.Update, db *gorm.DB, userID int64, s *registerState) {
	chatID := update.Message.Chat.ID
	code := strings.TrimSpace(update.Message.Text)

	refreshToken, err := gmail.ExchangeCode(ctx, s.oauthConfig, code)
	if err != nil {
		sendText(ctx, b, chatID, "❌ Failed to exchange authorization code: "+escapeMarkdown(err.Error())+"\n\nPlease run /register again\\.")
		setState(userID, nil)
		return
	}

	if err := gmail.SmokeTest(ctx, s.oauthConfig, refreshToken); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("register: gmail smoke test failed")
		sendText(ctx, b, chatID, "❌ Credentials are invalid or Gmail access was not granted\\. Please run /register again\\.")
		setState(userID, nil)
		return
	}
	logrus.WithField("user_id", userID).Info("register: gmail smoke test passed")

	if err := commands.UpsertCredentials(db, userID, s.oauthConfig.ClientID, s.oauthConfig.ClientSecret, refreshToken); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("register: failed to save refresh token")
		sendText(ctx, b, chatID, "❌ Internal error saving credentials\\. Please try again later\\.")
		setState(userID, nil)
		return
	}
	logrus.WithField("user_id", userID).Info("register: refresh token saved")

	if err := commands.SetRegistered(db, userID, true); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("register: failed to mark user as registered")
	} else {
		logrus.WithField("user_id", userID).Info("register: user marked as registered")
	}

	setState(userID, nil)

	_, err = b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    chatID,
		Text:      "✅ Gmail account linked successfully\\! You can now configure your email monitoring\\.",
		ParseMode: models.ParseModeMarkdown,
		ReplyMarkup: &models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{{Text: "⚙️ Configure"}},
			},
			ResizeKeyboard:  true,
			OneTimeKeyboard: false,
		},
	})
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("register: failed to send success message")
	}
}

func sendText(ctx context.Context, b *tgbot.Bot, chatID int64, text string) {
	_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		logrus.WithError(err).Error("sendText: failed to send message")
	}
}

// escapeMarkdown escapes characters that have special meaning in MarkdownV2.
func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]",
		"(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`",
		">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-",
		"=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}",
		".", "\\.", "!", "\\!",
	)
	return replacer.Replace(s)
}

const registerGuideMessage = `*Link your Gmail account*

Paste the full contents of your *credentials\.json* file \(from GCP\) into this chat\.

_You can get this file from Google Cloud Console → APIs & Services → Credentials → your Desktop app OAuth client\._`
