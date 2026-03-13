package addprompt

import (
	"context"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram/handlers"
)

// AddPromptHandler handles /addprompt — starts the new-prompt conversation flow.
func AddPromptHandler(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		startAddPromptFlow(ctx, bot, update.Message.From.ID, update.Message.Chat.ID, nil)
	}
}

// startAddPromptFlow initialises the addPrompt state and asks for the sender filter.
// editID is nil for new prompts, non-nil when editing an existing one.
func startAddPromptFlow(ctx context.Context, bot *tgbot.Bot, userID, chatID int64, editID *uuid.UUID) {
	setAddPromptState(userID, &addPromptState{step: stepWaitFilter, editID: editID})
	_, err := bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    chatID,
		Text:      "*Step 1/2* — Enter the sender filter \\(e\\.g\\. `newsletter@example\\.com`\\), or press the button below to match all senders\\.",
		ParseMode: models.ParseModeMarkdown,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "📭 No sender filter", CallbackData: "addprompt:nofilter"}},
			},
		},
	})
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("addprompt: failed to send filter prompt")
	}
}

// HandleAddPromptConversation routes text messages for users inside the addPrompt flow.
func HandleAddPromptConversation(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		s := getAddPromptState(userID)
		if s == nil {
			return
		}

		text := strings.TrimSpace(update.Message.Text)

		switch s.step {
		case stepWaitFilter:
			filter := text
			if filter == "-" {
				filter = ""
			}
			s.filter = filter
			s.step = stepWaitPrompt
			setAddPromptState(userID, s)
			handlers.SendText(ctx, bot, chatID, "*Step 2/2* — Enter the summarization prompt text\\.")

		case stepWaitPrompt:
			if text == "" {
				handlers.SendText(ctx, bot, chatID, "Prompt text cannot be empty\\. Please enter a prompt\\.")
				return
			}
			if err := savePrompt(ctx, svc, userID, s, text); err != nil {
				logrus.WithError(err).WithField("user_id", userID).Error("addprompt: failed to save prompt")
				handlers.SendText(ctx, bot, chatID, "Failed to save prompt\\. Please try again\\.")
				return
			}
			setAddPromptState(userID, nil)
			logrus.WithField("user_id", userID).Info("addprompt: prompt saved")
			handlers.SendText(ctx, bot, chatID, "✅ Prompt saved\\!")
			handlers.SendPromptList(ctx, bot, svc, userID, chatID)
		}
	}
}

func savePrompt(ctx context.Context, svc service.PromptService, userID int64, s *addPromptState, promptText string) error {
	if s.editID != nil {
		return svc.UpdatePrompt(ctx, *s.editID, promptText, s.filter)
	}
	return svc.AddPrompt(ctx, userID, promptText, s.filter)
}

// EditPromptCallback handles "edit:<uuid>" inline button presses.
func EditPromptCallback(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID

		idStr := strings.TrimPrefix(update.CallbackQuery.Data, "edit:")
		id, err := uuid.Parse(idStr)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("editprompt: invalid uuid in callback")
			handlers.AnswerCallback(ctx, bot, update.CallbackQuery.ID)
			return
		}

		handlers.AnswerCallback(ctx, bot, update.CallbackQuery.ID)
		startAddPromptFlow(ctx, bot, userID, chatID, &id)
	}
}

// RemovePromptCallback handles "remove:<uuid>" inline button presses.
func RemovePromptCallback(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID

		idStr := strings.TrimPrefix(update.CallbackQuery.Data, "remove:")
		id, err := uuid.Parse(idStr)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("removeprompt: invalid uuid in callback")
			handlers.AnswerCallback(ctx, bot, update.CallbackQuery.ID)
			return
		}

		if err := svc.DeletePrompt(ctx, id); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("removeprompt: failed to delete")
			handlers.AnswerCallback(ctx, bot, update.CallbackQuery.ID)
			handlers.SendText(ctx, bot, chatID, "Failed to remove prompt\\. Please try again\\.")
			return
		}

		logrus.WithField("user_id", userID).Info("removeprompt: prompt deleted")
		handlers.AnswerCallback(ctx, bot, update.CallbackQuery.ID)
		handlers.SendText(ctx, bot, chatID, "🗑 Prompt removed\\.")
		handlers.SendPromptList(ctx, bot, svc, userID, chatID)
	}
}

// AddPromptNoFilterCallback handles the "📭 No sender filter" inline button.
func AddPromptNoFilterCallback(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID
		handlers.AnswerCallback(ctx, bot, update.CallbackQuery.ID)

		s := getAddPromptState(userID)
		if s == nil || s.step != stepWaitFilter {
			return
		}
		s.filter = ""
		s.step = stepWaitPrompt
		setAddPromptState(userID, s)
		handlers.SendText(ctx, bot, chatID, "*Step 2/2* — Enter the summarization prompt text\\.")
	}
}

// AddPromptNewCallback handles the "➕ Add new" inline button.
func AddPromptNewCallback(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID
		handlers.AnswerCallback(ctx, bot, update.CallbackQuery.ID)
		startAddPromptFlow(ctx, bot, userID, chatID, nil)
	}
}
