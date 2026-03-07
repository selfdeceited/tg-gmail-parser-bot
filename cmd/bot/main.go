package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/telegram"
)

func main() {
	_ = godotenv.Load()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	_, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b, err := tgbot.New(token,
		tgbot.WithDefaultHandler(telegram.DefaultHandler),
	)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	telegram.RegisterHandlers(b)

	_, err = b.SetMyCommands(ctx, &tgbot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "start", Description: "Start the bot and see the setup guide"},
			{Command: "register", Description: "Link a Gmail account for monitoring"},
		},
	})
	if err != nil {
		log.Printf("warning: failed to set bot commands: %v", err)
	}

	log.Println("bot started, waiting for updates...")
	b.Start(ctx)
}
