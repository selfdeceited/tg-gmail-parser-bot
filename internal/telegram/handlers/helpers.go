package handlers

import (
	"context"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// SendText sends a MarkdownV2-formatted message to a chat.
func SendText(ctx context.Context, bot *tgbot.Bot, chatID int64, text string) {
	_, err := bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		logrus.WithError(err).Error("sendText: failed to send message")
	}
}

// AnswerCallback acknowledges an inline keyboard callback query.
func AnswerCallback(ctx context.Context, bot *tgbot.Bot, callbackID string) {
	_, err := bot.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{CallbackQueryID: callbackID})
	if err != nil {
		logrus.WithError(err).Error("failed to answer callback query")
	}
}

// EscapeMarkdown escapes characters that have special meaning in MarkdownV2.
func EscapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]",
		"(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`",
		">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-",
		"=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}",
		".", "\\.", "!", "\\!",
	)
	return replacer.Replace(s)
}

const notRegisteredMsg = "You need to /register before using this command\\."

// RequireRegistered wraps a message handler — passes through only if the user is registered.
func RequireRegistered(registrationService service.RegistrationService, next tgbot.HandlerFunc) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		userID := update.Message.From.ID
		ok, err := registrationService.IsRegistered(ctx, userID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("requireRegistered: failed to check registration")
			SendText(ctx, bot, update.Message.Chat.ID, "Something went wrong\\. Please try again\\.")
			return
		}
		if !ok {
			SendText(ctx, bot, update.Message.Chat.ID, notRegisteredMsg)
			return
		}
		next(ctx, bot, update)
	}
}

// RequireRegisteredCB wraps a callback query handler — passes through only if the user is registered.
func RequireRegisteredCB(registrationService service.RegistrationService, next tgbot.HandlerFunc) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID
		ok, err := registrationService.IsRegistered(ctx, userID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("requireRegisteredCB: failed to check registration")
			AnswerCallback(ctx, bot, update.CallbackQuery.ID)
			SendText(ctx, bot, chatID, "Something went wrong\\. Please try again\\.")
			return
		}
		if !ok {
			AnswerCallback(ctx, bot, update.CallbackQuery.ID)
			SendText(ctx, bot, chatID, notRegisteredMsg)
			return
		}
		next(ctx, bot, update)
	}
}
