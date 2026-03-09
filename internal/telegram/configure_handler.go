package telegram

import (
	"context"
	"fmt"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// ConfigureCommandHandler handles /configure — lists the user's prompts with inline Edit/Remove buttons.
func ConfigureCommandHandler(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		sendPromptList(ctx, bot, svc, userID, chatID)
	}
}

// sendPromptList fetches and displays the current prompt list for a user.
// Shared by ConfigureCommandHandler and post-edit/delete refresh.
func sendPromptList(ctx context.Context, bot *tgbot.Bot, svc service.PromptService, userID, chatID int64) {
	prompts, err := svc.ListPrompts(ctx, userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("configure: failed to list prompts")
		sendText(ctx, bot, chatID, "Failed to load prompts\\. Please try again\\.")
		return
	}

	if len(prompts) == 0 {
		_, err := bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      "You have no prompts yet\\. Use /addprompt to add one\\.",
			ParseMode: models.ParseModeMarkdown,
			ReplyMarkup: &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{{Text: "➕ Add new", CallbackData: "addprompt:new"}},
				},
			},
		})
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("configure: failed to send empty state message")
		}
		return
	}

	for _, p := range prompts {
		filterLine := "_none_"
		if p.Filter != "" {
			filterLine = escapeMarkdown(p.Filter)
		}
		text := fmt.Sprintf("📌 *Filter:* %s\n💬 %s", filterLine, escapeMarkdown(p.Prompt))
		idStr := p.ID.String()
		_, err := bot.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
			ReplyMarkup: &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						{Text: "✏️ Edit", CallbackData: "edit:" + idStr},
						{Text: "🗑 Remove", CallbackData: "remove:" + idStr},
					},
				},
			},
		})
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("configure: failed to send prompt message")
		}
	}

	_, err = bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    chatID,
		Text:      "Your prompts are listed above\\.",
		ParseMode: models.ParseModeMarkdown,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "➕ Add new", CallbackData: "addprompt:new"}},
			},
		},
	})
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("configure: failed to send footer message")
	}
}
