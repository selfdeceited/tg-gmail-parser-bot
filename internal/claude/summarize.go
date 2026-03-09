package claude

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/gmail"
)

const model = anthropic.ModelClaudeHaiku4_5_20251001

// SummarizeResult holds the structured response from Claude.
type SummarizeResult struct {
	Result  string `json:"result"`  // "matched" | "not matched"
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Client wraps the Anthropic SDK for email summarization.
type Client struct {
	inner *anthropic.Client
}

// NewClient creates a Client using the provided API key.
func NewClient(apiKey string) *Client {
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Client{inner: &c}
}

// Summarize sends the email through Claude using the augmented prompt and
// returns a structured result. Returns an error if the API call fails or
// the response cannot be parsed.
func (c *Client) Summarize(ctx context.Context, userPrompt string, email gmail.EmailMessage) (*SummarizeResult, error) {
	augmented := buildPrompt(userPrompt, email)

	logrus.WithFields(logrus.Fields{
		"message_id": email.ID,
		"subject":    email.Subject,
		"from":       email.From,
	}).Info("claude: sending message for summarization")

	msg, err := c.inner.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(augmented)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	if len(msg.Content) == 0 {
		return nil, fmt.Errorf("claude returned empty response")
	}

	raw := msg.Content[0].Text
	logrus.WithFields(logrus.Fields{
		"message_id": email.ID,
		"response":   raw,
	}).Info("claude: received response")

	var result SummarizeResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("failed to parse claude response as JSON: %w (raw: %s)", err, raw)
	}

	logrus.WithFields(logrus.Fields{
		"message_id": email.ID,
		"result":     result.Result,
		"title":      result.Title,
	}).Info("claude: summarization complete")

	return &result, nil
}

func buildPrompt(userPrompt string, email gmail.EmailMessage) string {
	return fmt.Sprintf(
		`Identify if the email matches the stated criteria in the prompt. If not, result should be "not matched".

Prompt: %s

Email subject: %s
Email from: %s
Email body:
%s

Answer in JSON format with the following fields:
- result: "matched" or "not matched"
- title: email subject/title
- content: summary of the email content per the prompt criteria`,
		userPrompt,
		email.Subject,
		email.From,
		email.Body,
	)
}
