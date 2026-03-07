package telegram

import (
	"context"
	"log"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// RegisterHandler is a placeholder for the /register command.
func RegisterHandler(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      "⚙️ /register is not implemented yet\\.",
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Printf("RegisterHandler: failed to send message: %v", err)
	}
}
