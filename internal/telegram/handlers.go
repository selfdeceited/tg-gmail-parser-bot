package telegram

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// RegisterHandlers registers all bot command and callback handlers.
func RegisterHandlers(bot *tgbot.Bot, registrationService service.RegistrationService, promptService service.PromptService, watchService service.WatchService) {
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact, StartHandler)
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/register", tgbot.MatchTypeExact, RegisterHandler(registrationService))
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/clearregistration", tgbot.MatchTypeExact, ClearRegistrationHandler(registrationService))
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/configure", tgbot.MatchTypeExact, requireRegistered(registrationService, ConfigureCommandHandler(promptService)))
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/addprompt", tgbot.MatchTypeExact, requireRegistered(registrationService, AddPromptHandler(promptService)))
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "/watch", tgbot.MatchTypeExact, requireRegistered(registrationService, WatchHandler(watchService)))
	bot.RegisterHandler(tgbot.HandlerTypeMessageText, "⚙️ Configure", tgbot.MatchTypeExact, requireRegistered(registrationService, ConfigureCommandHandler(promptService)))

	bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "edit:", tgbot.MatchTypePrefix, requireRegisteredCB(registrationService, EditPromptCallback(promptService)))
	bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "remove:", tgbot.MatchTypePrefix, requireRegisteredCB(registrationService, RemovePromptCallback(promptService)))
	bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "addprompt:new", tgbot.MatchTypeExact, requireRegisteredCB(registrationService, AddPromptNewCallback(promptService)))
	bot.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "addprompt:nofilter", tgbot.MatchTypeExact, requireRegisteredCB(registrationService, AddPromptNoFilterCallback(promptService)))

	logrus.Info("handlers registered")
}

// DefaultHandler routes non-command messages to active conversation flows.
func DefaultHandler(registrationService service.RegistrationService, promptService service.PromptService) tgbot.HandlerFunc {
	registerConv := HandleConversation(registrationService)
	addPromptConv := HandleAddPromptConversation(promptService)
	return func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
		registerConv(ctx, bot, update)
		addPromptConv(ctx, bot, update)
	}
}
