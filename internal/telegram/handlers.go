package telegram

import (
	"context"
	"log"

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

// StartHandler responds to /start with the GCP OAuth setup guide.
func StartHandler(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      startMessage,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Printf("StartHandler: failed to send message: %v", err)
	}
}

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

// <!-- REVIEW: spec://common/main#formatting says "Markdown parse mode" but
// ParseModeMarkdown (legacy) is deprecated in the Bot API and fails on special
// characters. Using ParseModeMarkdownV2 instead. -->

const startMessage = `*Welcome to tg\-gmail\-parser\-bot\!* 👋

This bot monitors your Gmail inbox and forwards parsed email summaries to Telegram\.

*Setup Guide — GCP OAuth Token*

To authorize Gmail access you need a GCP OAuth 2\.0 client credential\. Follow these steps:

*Step 1 — Create a GCP project*
1\. Go to [Google Cloud Console](https://console\.cloud\.google\.com/)
2\. Create a new project \(or select an existing one\)

*Step 2 — Enable the Gmail API*
1\. In your project, go to *APIs & Services → Library*
2\. Search for "Gmail API" and click *Enable*

*Step 3 — Create OAuth 2\.0 credentials*
1\. Go to *APIs & Services → Credentials*
2\. Click *Create Credentials → OAuth client ID*
3\. Select application type: *Desktop app*
4\. Download the generated *credentials\.json* file

*Step 4 — Link your Gmail account*
Once you have your *credentials\.json*, use /register to connect your Gmail account\.`
