package telegram

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// RegisterHandlers registers all bot command handlers.
func RegisterHandlers(b *tgbot.Bot) {
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact, StartHandler)
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/register", tgbot.MatchTypeExact, RegisterHandler)
}

// DefaultHandler silently ignores unrecognized messages.
func DefaultHandler(_ context.Context, _ *tgbot.Bot, _ *models.Update) {}
