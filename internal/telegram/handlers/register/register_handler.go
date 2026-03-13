package register

import (
	"context"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/gmail"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram/handlers"
)

// RegisterConversationHandler handles the /register command (step 1: ask for credentials.json).
func RegisterConversationHandler(svc service.RegistrationService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		if err := svc.VerifyCredentials(ctx, userID); err == nil {
			logrus.WithField("user_id", userID).Info("register: already registered with valid credentials, skipping")
			_, _ = bot.SendMessage(ctx, &tgbot.SendMessageParams{
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
		handlers.SendText(ctx, bot, chatID, registerGuideMessage)
	}
}

// HandleRegisterConversation routes non-command messages for users inside the register flow.
func HandleRegisterConversation(svc service.RegistrationService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
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
			handleCredentialsPaste(ctx, bot, update, userID, s)
		case stepWaitAuthCode:
			handleAuthCodePaste(ctx, bot, update, svc, userID, s)
		}
	}
}

func handleCredentialsPaste(ctx context.Context, bot *tgbot.Bot, update *models.Update, userID int64, s *registerState) {
	chatID := update.Message.Chat.ID
	raw := strings.TrimSpace(update.Message.Text)

	cfg, err := gmail.ParseCredentials([]byte(raw))
	if err != nil {
		handlers.SendText(ctx, bot, chatID, "❌ Could not parse credentials\\.json: "+handlers.EscapeMarkdown(err.Error())+"\n\nPlease paste the full contents of your credentials\\.json file\\.")
		return
	}

	authURL := gmail.BuildAuthURL(cfg)
	s.oauthConfig = cfg
	s.step = stepWaitAuthCode
	setState(userID, s)

	handlers.SendText(ctx, bot, chatID,
		"✅ Credentials parsed\\.\n\n"+
			"*Step 2 — Authorize Gmail access*\n\n"+
			"Visit the link below, sign in with your Google account, and paste the authorization code here:\n\n"+
			"`"+handlers.EscapeMarkdown(authURL)+"`",
	)
}

func handleAuthCodePaste(ctx context.Context, bot *tgbot.Bot, update *models.Update, svc service.RegistrationService, userID int64, s *registerState) {
	chatID := update.Message.Chat.ID
	code := strings.TrimSpace(update.Message.Text)

	refreshToken, err := gmail.ExchangeCode(ctx, s.oauthConfig, code)
	if err != nil {
		handlers.SendText(ctx, bot, chatID, "❌ Failed to exchange authorization code: "+handlers.EscapeMarkdown(err.Error())+"\n\nPlease run /register again\\.")
		setState(userID, nil)
		return
	}

	if err := gmail.SmokeTest(ctx, s.oauthConfig, refreshToken); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("register: gmail smoke test failed")
		handlers.SendText(ctx, bot, chatID, "❌ Credentials are invalid or Gmail access was not granted\\. Please run /register again\\.")
		setState(userID, nil)
		return
	}
	logrus.WithField("user_id", userID).Info("register: gmail smoke test passed")

	if err := svc.SaveCredentials(ctx, userID, s.oauthConfig.ClientID, s.oauthConfig.ClientSecret, refreshToken); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("register: failed to save credentials")
		handlers.SendText(ctx, bot, chatID, "❌ Internal error saving credentials\\. Please try again later\\.")
		setState(userID, nil)
		return
	}
	logrus.WithField("user_id", userID).Info("register: credentials saved, user registered")

	setState(userID, nil)

	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
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

const registerGuideMessage = `*Link your Gmail account*

Paste the full contents of your *credentials\.json* file \(from GCP\) into this chat\.

_You can get this file from Google Cloud Console → APIs & Services → Credentials → your Desktop app OAuth client\._`
