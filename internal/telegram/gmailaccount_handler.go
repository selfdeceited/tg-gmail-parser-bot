package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// SetGmailAccountCallback handles the "gmailaccount:set" inline button — starts the account index flow.
func SetGmailAccountCallback(svc service.RegistrationService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID
		answerCallback(ctx, bot, update.CallbackQuery.ID)

		setSetGmailAccountState(userID, &setGmailAccountState{})
		sendText(ctx, bot, chatID,
			"Enter your Gmail account index \\(the number `N` in `/u/N/` of Gmail URLs\\)\\.\n"+
				"Use `0` for your primary account, `1` for the second, etc\\.")
	}
}

// HandleSetGmailAccountConversation routes text input for users in the setGmailAccount flow.
func HandleSetGmailAccountConversation(regSvc service.RegistrationService, promptSvc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		if getSetGmailAccountState(userID) == nil {
			return
		}

		text := strings.TrimSpace(update.Message.Text)
		index, err := strconv.Atoi(text)
		if err != nil || index < 0 {
			sendText(ctx, bot, chatID, "Please enter a non\\-negative integer \\(e\\.g\\. `0`, `1`, `2`\\)\\.")
			return
		}

		if err := regSvc.SetGmailAccountIndex(ctx, userID, index); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("gmailaccount: failed to save index")
			sendText(ctx, bot, chatID, "Failed to save Gmail account index\\. Please try again\\.")
			return
		}

		setSetGmailAccountState(userID, nil)
		logrus.WithFields(logrus.Fields{"user_id": userID, "index": index}).Info("gmailaccount: index updated")
		sendText(ctx, bot, chatID, fmt.Sprintf("✅ Gmail account index set to `%d`\\.", index))
		sendAccountSettings(ctx, bot, regSvc, userID, chatID)
	}
}

// sendAccountSettings sends the Gmail account index card with a change button.
func sendAccountSettings(ctx context.Context, bot *tgbot.Bot, svc service.RegistrationService, userID, chatID int64) {
	index, err := svc.GetGmailAccountIndex(ctx, userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("configure: failed to get gmail account index")
		return
	}

	text := fmt.Sprintf("🔗 *Gmail account index:* `%d`\nLinks will open `mail\\.google\\.com/mail/u/%d/`", index, index)
	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "✏️ Change account index", CallbackData: "gmailaccount:set"}},
			},
		},
	})
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("configure: failed to send account settings")
	}
}
