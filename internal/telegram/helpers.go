package telegram

import (
	"context"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

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

const notRegisteredMsg = "You need to /register before using this command\\."

// requireRegistered wraps a message handler — passes through only if the user is registered.
func requireRegistered(regSvc service.RegistrationService, next tgbot.HandlerFunc) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		userID := update.Message.From.ID
		ok, err := regSvc.IsRegistered(ctx, userID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("requireRegistered: failed to check registration")
			sendText(ctx, b, update.Message.Chat.ID, "Something went wrong\\. Please try again\\.")
			return
		}
		if !ok {
			sendText(ctx, b, update.Message.Chat.ID, notRegisteredMsg)
			return
		}
		next(ctx, b, update)
	}
}

// requireRegisteredCB wraps a callback query handler — passes through only if the user is registered.
func requireRegisteredCB(regSvc service.RegistrationService, next tgbot.HandlerFunc) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID
		ok, err := regSvc.IsRegistered(ctx, userID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("requireRegisteredCB: failed to check registration")
			answerCallback(ctx, b, update.CallbackQuery.ID)
			sendText(ctx, b, chatID, "Something went wrong\\. Please try again\\.")
			return
		}
		if !ok {
			answerCallback(ctx, b, update.CallbackQuery.ID)
			sendText(ctx, b, chatID, notRegisteredMsg)
			return
		}
		next(ctx, b, update)
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
