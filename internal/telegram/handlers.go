package telegram

import (
	tgbot "github.com/go-telegram/bot"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/service"
)

// RegisterHandlers registers all bot command handlers.
func RegisterHandlers(b *tgbot.Bot, svc service.RegistrationService) {
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact, StartHandler)
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/register", tgbot.MatchTypeExact, RegisterHandler(svc))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/clearregistration", tgbot.MatchTypeExact, ClearRegistrationHandler(svc))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "⚙️ Configure", tgbot.MatchTypeExact, ConfigureButtonHandler(svc))
	logrus.Info("handlers registered")
}

// DefaultHandler routes non-command messages to active conversation flows.
func DefaultHandler(svc service.RegistrationService) tgbot.HandlerFunc {
	return HandleConversation(svc)
}
