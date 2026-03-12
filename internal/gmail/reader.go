package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmailapi "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// EmailMessage holds the parsed fields of a Gmail message.
type EmailMessage struct {
	ID      string
	Subject string
	From    string
	Body    string
	URL     string
}

// NewGmailService builds an authenticated Gmail API service from stored credentials.
func NewGmailService(ctx context.Context, clientID, clientSecret, refreshToken string) (*gmailapi.Service, error) {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       gmailScopes,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}
	ts := cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	svc, err := gmailapi.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("failed to create gmail service: %w", err)
	}
	return svc, nil
}

// FetchNewMessages returns all messages received after the given timestamp.
// accountIndex is the Gmail /u/<N>/ slot used to build direct message URLs.
func FetchNewMessages(ctx context.Context, svc *gmailapi.Service, since time.Time, accountIndex int) ([]EmailMessage, error) {
	query := fmt.Sprintf("after:%d", since.Unix())
	logrus.WithFields(logrus.Fields{"query": query}).Info("gmail: fetching new messages")

	list, err := svc.Users.Messages.List("me").Q(query).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	if len(list.Messages) == 0 {
		logrus.Info("gmail: no new messages")
		return nil, nil
	}

	logrus.WithField("count", len(list.Messages)).Info("gmail: new messages found")

	var emails []EmailMessage
	for _, m := range list.Messages {
		msg, err := svc.Users.Messages.Get("me", m.Id).Format("full").Do()
		if err != nil {
			logrus.WithError(err).WithField("message_id", m.Id).Warn("gmail: failed to fetch message, skipping")
			continue
		}
		email := parseMessage(msg, accountIndex)
		logrus.WithFields(logrus.Fields{
			"message_id": email.ID,
			"from":       email.From,
			"subject":    email.Subject,
		}).Info("gmail: message parsed")
		emails = append(emails, email)
	}
	return emails, nil
}

func parseMessage(msg *gmailapi.Message, accountIndex int) EmailMessage {
	email := EmailMessage{
		ID:   msg.Id,
		URL:  fmt.Sprintf("https://mail.google.com/mail/u/%d/#all/%s", accountIndex, msg.Id),
		Body: msg.Snippet,
	}

	for _, h := range msg.Payload.Headers {
		switch strings.ToLower(h.Name) {
		case "subject":
			email.Subject = h.Value
		case "from":
			email.From = h.Value
		}
	}

	if body := extractBody(msg.Payload); body != "" {
		email.Body = body
	}

	return email
}

// extractBody walks the message payload tree looking for a text/plain part.
func extractBody(part *gmailapi.MessagePart) string {
	if part == nil {
		return ""
	}
	if strings.HasPrefix(part.MimeType, "text/plain") {
		data, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err == nil && len(data) > 0 {
			return string(data)
		}
	}
	for _, sub := range part.Parts {
		if body := extractBody(sub); body != "" {
			return body
		}
	}
	return ""
}
