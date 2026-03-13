package telegram

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram/handlers"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram/handlers/addprompt"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram/handlers/gmailaccount"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram/handlers/register"
)

// RegisterHandlers registers all bot command and callback handlers.
func RegisterHandlers(bot *tgbot.Bot, registrationService service.RegistrationService, promptService service.PromptService, watchService service.WatchService) {
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact,
		handlers.StartHandler)
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/register", tgbot.MatchTypeExact,
		register.RegisterConversationHandler(registrationService))
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/clearregistration",
		tgbot.MatchTypeExact, handlers.ClearRegistrationHandler(registrationService))
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/configure", tgbot.MatchTypeExact,
		handlers.RequireRegistered(registrationService, handlers.ConfigureCommandHandler(promptService, registrationService)))
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/addprompt", tgbot.MatchTypeExact,
		handlers.RequireRegistered(registrationService, addprompt.AddPromptHandler(promptService)))
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/watch", tgbot.MatchTypeExact,
		handlers.RequireRegistered(registrationService, handlers.WatchHandler(watchService)))
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "⚙️ Configure", tgbot.MatchTypeExact,
		handlers.RequireRegistered(registrationService, handlers.ConfigureCommandHandler(promptService, registrationService)))

	bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "edit:", tgbot.MatchTypePrefix,
		handlers.RequireRegisteredCB(registrationService, addprompt.EditPromptCallback(promptService)))
	bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "remove:", tgbot.MatchTypePrefix,
		handlers.RequireRegisteredCB(registrationService, addprompt.RemovePromptCallback(promptService)))
	bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "addprompt:new", tgbot.MatchTypeExact,
		handlers.RequireRegisteredCB(registrationService, addprompt.AddPromptNewCallback(promptService)))
	bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "addprompt:nofilter", tgbot.MatchTypeExact,
		handlers.RequireRegisteredCB(registrationService, addprompt.AddPromptNoFilterCallback(promptService)))
	bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "gmailaccount:set", tgbot.MatchTypeExact, handlers.RequireRegisteredCB(registrationService, gmailaccount.SetGmailAccountCallback(registrationService)))

	logrus.Info("handlers registered")
}

// StartBot registers bot commands and starts polling. Blocks until ctx is cancelled.
func StartBot(ctx context.Context, bot *tgbot.Bot) {
	_, err := bot.SetMyCommands(ctx, &tgbot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "start", Description: "Start the bot and see the setup guide"},
			{Command: "register", Description: "Link a Gmail account for monitoring"},
			{Command: "clearregistration", Description: "Unlink Gmail account and clear credentials"},
			{Command: "configure", Description: "Manage summarization prompts"},
			{Command: "addprompt", Description: "Add or edit a summarization prompt"},
			{Command: "watch", Description: "Start or stop Gmail monitoring"},
		},
	})
	if err != nil {
		logrus.WithError(err).Warn("failed to set bot commands")
	}

	logrus.Info("bot started, waiting for updates...")
	bot.Start(ctx)
}

// ConversationHandler routes non-command messages to active conversation flows.
func ConversationHandler(registrationService service.RegistrationService, promptService service.PromptService) tgbot.HandlerFunc {
	registerConv := register.HandleRegisterConversation(registrationService)
	addPromptConv := addprompt.HandleAddPromptConversation(promptService)
	gmailAccountConv := gmailaccount.HandleSetGmailAccountConversation(registrationService)
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		registerConv(ctx, bot, update)
		addPromptConv(ctx, bot, update)
		gmailAccountConv(ctx, bot, update)
	}
}
