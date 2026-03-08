package telegram

import (
	"context"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// AddPromptHandler handles /addprompt — starts the new-prompt conversation flow.
func AddPromptHandler(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		startAddPromptFlow(ctx, b, update.Message.From.ID, update.Message.Chat.ID, nil)
	}
}

// startAddPromptFlow initialises the addPrompt state and asks for the sender filter.
// editID is nil for new prompts, non-nil when editing an existing one.
func startAddPromptFlow(ctx context.Context, b *tgbot.Bot, userID, chatID int64, editID *uuid.UUID) {
	setAddPromptState(userID, &addPromptState{step: stepAddPromptWaitFilter, editID: editID})
	_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
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
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
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
		case stepAddPromptWaitFilter:
			filter := text
			if filter == "-" {
				filter = ""
			}
			s.filter = filter
			s.step = stepAddPromptWaitPrompt
			setAddPromptState(userID, s)
			sendText(ctx, b, chatID, "*Step 2/2* — Enter the summarization prompt text\\.")

		case stepAddPromptWaitPrompt:
			if text == "" {
				sendText(ctx, b, chatID, "Prompt text cannot be empty\\. Please enter a prompt\\.")
				return
			}
			if err := savePrompt(ctx, svc, userID, s, text); err != nil {
				logrus.WithError(err).WithField("user_id", userID).Error("addprompt: failed to save prompt")
				sendText(ctx, b, chatID, "Failed to save prompt\\. Please try again\\.")
				return
			}
			setAddPromptState(userID, nil)
			logrus.WithField("user_id", userID).Info("addprompt: prompt saved")
			sendText(ctx, b, chatID, "✅ Prompt saved\\!")
			sendPromptList(ctx, b, svc, userID, chatID)
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
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID

		idStr := strings.TrimPrefix(update.CallbackQuery.Data, "edit:")
		id, err := uuid.Parse(idStr)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("editprompt: invalid uuid in callback")
			answerCallback(ctx, b, update.CallbackQuery.ID)
			return
		}

		answerCallback(ctx, b, update.CallbackQuery.ID)
		startAddPromptFlow(ctx, b, userID, chatID, &id)
	}
}

// RemovePromptCallback handles "remove:<uuid>" inline button presses.
func RemovePromptCallback(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID

		idStr := strings.TrimPrefix(update.CallbackQuery.Data, "remove:")
		id, err := uuid.Parse(idStr)
		if err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("removeprompt: invalid uuid in callback")
			answerCallback(ctx, b, update.CallbackQuery.ID)
			return
		}

		if err := svc.DeletePrompt(ctx, id); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Error("removeprompt: failed to delete")
			answerCallback(ctx, b, update.CallbackQuery.ID)
			sendText(ctx, b, chatID, "Failed to remove prompt\\. Please try again\\.")
			return
		}

		logrus.WithField("user_id", userID).Info("removeprompt: prompt deleted")
		answerCallback(ctx, b, update.CallbackQuery.ID)
		sendText(ctx, b, chatID, "🗑 Prompt removed\\.")
		sendPromptList(ctx, b, svc, userID, chatID)
	}
}

// AddPromptNoFilterCallback handles the "📭 No sender filter" inline button (callback: "addprompt:nofilter").
func AddPromptNoFilterCallback(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID
		answerCallback(ctx, b, update.CallbackQuery.ID)

		s := getAddPromptState(userID)
		if s == nil || s.step != stepAddPromptWaitFilter {
			return
		}
		s.filter = ""
		s.step = stepAddPromptWaitPrompt
		setAddPromptState(userID, s)
		sendText(ctx, b, chatID, "*Step 2/2* — Enter the summarization prompt text\\.")
	}
}

// AddPromptNewCallback handles the "➕ Add new" inline button (callback: "addprompt:new").
func AddPromptNewCallback(svc service.PromptService) tgbot.HandlerFunc {
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID
		chatID := update.CallbackQuery.Message.Message.Chat.ID
		answerCallback(ctx, b, update.CallbackQuery.ID)
		startAddPromptFlow(ctx, b, userID, chatID, nil)
	}
}

func answerCallback(ctx context.Context, b *tgbot.Bot, callbackID string) {
	_, err := b.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{CallbackQueryID: callbackID})
	if err != nil {
		logrus.WithError(err).Error("failed to answer callback query")
	}
}
