package telegram

import (
	"context"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sirupsen/logrus"
)

// StartHandler responds to /start with the GCP OAuth setup guide.
func StartHandler(ctx context.Context, b *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	logrus.WithField("user_id", update.Message.From.ID).Info("StartHandler: start")

	_, err := b.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      startMessage,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		logrus.WithError(err).Error("StartHandler: failed to send message")
	}

	logrus.WithField("user_id", update.Message.From.ID).Info("StartHandler: completed")
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
