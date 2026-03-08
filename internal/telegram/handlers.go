package telegram

import (
	tgbot "github.com/go-telegram/bot"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RegisterHandlers registers all bot command handlers.
func RegisterHandlers(b *tgbot.Bot, db *gorm.DB) {
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact, StartHandler)
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/register", tgbot.MatchTypeExact, RegisterHandler(db))
	b.RegisterHandler(tgbot.HandlerTypeMessageText, "⚙️ Configure", tgbot.MatchTypeExact, ConfigureButtonHandler(db))
	logrus.Info("handlers registered")
}

// DefaultHandler routes non-command messages to active conversation flows.
func DefaultHandler(db *gorm.DB) tgbot.HandlerFunc {
	return HandleConversation(db)
}
