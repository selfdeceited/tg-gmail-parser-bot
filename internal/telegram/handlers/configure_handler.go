package handlers

import (
	"context"
	"fmt"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// ConfigureCommandHandler handles /configure — shows account settings and lists prompts.
func ConfigureCommandHandler(promptSvc service.PromptService, regSvc service.RegistrationService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		SendAccountSettings(ctx, bot, regSvc, userID, chatID)
		SendPromptList(ctx, bot, promptSvc, userID, chatID)
	}
}

// SendPromptList fetches and displays the current prompt list for a user.
// Exported so the addprompt package can refresh the list after edits.
func SendPromptList(ctx context.Context, bot *tgbot.Bot, svc service.PromptService, userID, chatID int64) {
	prompts, err := svc.ListPrompts(ctx, userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("configure: failed to list prompts")
		SendText(ctx, bot, chatID, "Failed to load prompts\\. Please try again\\.")
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
			filterLine = EscapeMarkdown(p.Filter)
		}
		text := fmt.Sprintf("📌 *Filter:* %s\n💬 %s", filterLine, EscapeMarkdown(p.Prompt))
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

// SendAccountSettings sends the Gmail account index card with a change button.
// Exported so the gmailaccount package can refresh it after an update.
func SendAccountSettings(ctx context.Context, bot *tgbot.Bot, svc service.RegistrationService, userID, chatID int64) {
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
