package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const defaultIOTimeout = 10 * time.Second

type config struct {
	TelegramBotToken          string
	ClaudeAPIKey              string
	TokenEncryptionKey1       string
	TokenEncryptionKeyCurrent string
	IOTimeout                 time.Duration
}

// loadConfig loads .env (if present) and validates all required environment
// variables. Exits with code 1 if any are missing.
func loadConfig() config {
	_ = godotenv.Load()

	required := []string{
		"TELEGRAM_BOT_TOKEN",
		"CLAUDE_API_KEY",
		"TOKEN_ENCRYPTION_KEY_1",
		"TOKEN_ENCRYPTION_KEY_CURRENT",
	}

	var missing []string
	for _, key := range required {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		logrus.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
		fmt.Fprintf(os.Stderr, "Set the above variables in .env or the process environment.\n")
		os.Exit(1)
	}

	ioTimeout := defaultIOTimeout
	if raw := os.Getenv("IO_TIMEOUT"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			ioTimeout = parsed
		} else {
			logrus.Warnf("invalid IO_TIMEOUT %q, using default %s", raw, defaultIOTimeout)
		}
	}

	return config{
		TelegramBotToken:          os.Getenv("TELEGRAM_BOT_TOKEN"),
		ClaudeAPIKey:              os.Getenv("CLAUDE_API_KEY"),
		TokenEncryptionKey1:       os.Getenv("TOKEN_ENCRYPTION_KEY_1"),
		TokenEncryptionKeyCurrent: os.Getenv("TOKEN_ENCRYPTION_KEY_CURRENT"),
		IOTimeout:                 ioTimeout,
	}
}
