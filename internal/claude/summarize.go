package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/sirupsen/logrus"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/gmail"
)

const model = anthropic.ModelClaudeHaiku4_5_20251001

// SummarizeResult holds the structured response from Claude.
// Content is RawMessage to tolerate Claude returning either a plain string
// or a nested object; use ContentString() to get a displayable value.
type SummarizeResult struct {
	Result  string          `json:"result"` // "matched" | "not matched"
	Title   string          `json:"title"`
	Content json.RawMessage `json:"content"`
}

// ContentString returns content as a plain string regardless of whether
// Claude returned a JSON string or a JSON object.
func (r *SummarizeResult) ContentString() string {
	if len(r.Content) == 0 {
		return ""
	}
	// Try to unquote a JSON string first.
	var s string
	if err := json.Unmarshal(r.Content, &s); err == nil {
		return s
	}
	// Fall back to pretty-printed JSON for objects/arrays.
	var v any
	if err := json.Unmarshal(r.Content, &v); err == nil {
		b, _ := json.MarshalIndent(v, "", "  ")
		return string(b)
	}
	return string(r.Content)
}

// Client wraps the Anthropic SDK for email summarization.
type Client struct {
	inner *anthropic.Client
}

// NewClient creates a Client using the provided API key.
func NewClient(apiKey string) *Client {
	innerClient := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Client{inner: &innerClient}
}

// Summarize sends the email through Claude using the augmented prompt and
// returns a structured result. Returns an error if the API call fails or
// the response cannot be parsed.
func (client *Client) Summarize(ctx context.Context, userPrompt string, email gmail.EmailMessage) (*SummarizeResult, error) {
	augmented := buildPrompt(userPrompt, email)

	logrus.WithFields(logrus.Fields{
		"message_id": email.ID,
		"subject":    email.Subject,
		"from":       email.From,
	}).Info("claude: sending message for summarization")

	msg, err := client.inner.Messages.New(ctx, anthropic.MessageNewParams{
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

	raw := stripCodeFence(msg.Content[0].Text)
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

Respond with ONLY valid JSON — no markdown, no code fences, no extra text. The JSON must have exactly these fields:
- result: "matched" or "not matched"
- title: email subject/title (plain string)
- content: a plain text summary string of the email content per the prompt criteria — must be a JSON string, not an object`,
		userPrompt,
		email.Subject,
		email.From,
		email.Body,
	)
}

// stripCodeFence removes optional ```json ... ``` or ``` ... ``` wrapping from Claude responses.
func stripCodeFence(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		text = text[3:]
		if strings.HasPrefix(text, "json") {
			text = text[4:]
		}
		if idx := strings.LastIndex(text, "```"); idx != -1 {
			text = text[:idx]
		}
	}
	return strings.TrimSpace(text)
}
