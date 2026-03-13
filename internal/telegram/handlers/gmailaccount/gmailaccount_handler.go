package gmailaccount

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram/handlers"
)

// SetGmailAccountCallback handles the "gmailaccount:set" inline button — starts the account index flow.
func SetGmailAccountCallback(svc service.RegistrationService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID
		handlers.AnswerCallback(ctx, bot, update.CallbackQuery.ID)

		setSetGmailAccountState(userID, &setGmailAccountState{})
		handlers.SendText(ctx, bot, chatID,
			"Enter your Gmail account index \\(the number `N` in `/u/N/` of Gmail URLs\\)\\.\n"+
				"Use `0` for your primary account, `1` for the second, etc\\.")
	}
}

// HandleSetGmailAccountConversation routes text input for users in the setGmailAccount flow.
func HandleSetGmailAccountConversation(regSvc service.RegistrationService) tgbot.HandlerFunc {
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
			handlers.SendText(ctx, bot, chatID, "Please enter a non\\-negative integer \\(e\\.g\\. `0`, `1`, `2`\\)\\.")
			return
		}

		if err := regSvc.SetGmailAccountIndex(ctx, userID, index); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("gmailaccount: failed to save index")
			handlers.SendText(ctx, bot, chatID, "Failed to save Gmail account index\\. Please try again\\.")
			return
		}

		setSetGmailAccountState(userID, nil)
		logrus.WithFields(logrus.Fields{"user_id": userID, "index": index}).Info("gmailaccount: index updated")
		handlers.SendText(ctx, bot, chatID, fmt.Sprintf("✅ Gmail account index set to `%d`\\.", index))
		handlers.SendAccountSettings(ctx, bot, regSvc, userID, chatID)
	}
}
