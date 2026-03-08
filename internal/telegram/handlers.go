package telegram

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// RegisterHandlers registers all bot command and callback handlers.
func RegisterHandlers(b *tgbot.Bot, regSvc service.RegistrationService, promptSvc service.PromptService) {
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact, StartHandler)
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/register", tgbot.MatchTypeExact, RegisterHandler(regSvc))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/clearregistration", tgbot.MatchTypeExact, ClearRegistrationHandler(regSvc))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/configure", tgbot.MatchTypeExact, requireRegistered(regSvc, ConfigureCommandHandler(promptSvc)))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/addprompt", tgbot.MatchTypeExact, requireRegistered(regSvc, AddPromptHandler(promptSvc)))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "⚙️ Configure", tgbot.MatchTypeExact, requireRegistered(regSvc, ConfigureCommandHandler(promptSvc)))

	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "edit:", tgbot.MatchTypePrefix, requireRegisteredCB(regSvc, EditPromptCallback(promptSvc)))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "remove:", tgbot.MatchTypePrefix, requireRegisteredCB(regSvc, RemovePromptCallback(promptSvc)))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "addprompt:new", tgbot.MatchTypeExact, requireRegisteredCB(regSvc, AddPromptNewCallback(promptSvc)))
	b.RegisterHandler(tgbot.HandlerTypeCallbackQueryData, "addprompt:nofilter", tgbot.MatchTypeExact, requireRegisteredCB(regSvc, AddPromptNoFilterCallback(promptSvc)))

	logrus.Info("handlers registered")
}

// DefaultHandler routes non-command messages to active conversation flows.
func DefaultHandler(regSvc service.RegistrationService, promptSvc service.PromptService) tgbot.HandlerFunc {
	registerConv := HandleConversation(regSvc)
	addPromptConv := HandleAddPromptConversation(promptSvc)
	return func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		registerConv(ctx, b, update)
		addPromptConv(ctx, b, update)
	}
}
