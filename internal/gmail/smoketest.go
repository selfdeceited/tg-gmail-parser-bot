package gmail

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmailapi "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// VerifyRefreshToken runs a smoke test using stored client credentials and a refresh token.
func VerifyRefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) error {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       gmailScopes,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}
	return SmokeTest(ctx, cfg, refreshToken)
}

// SmokeTest verifies that the refresh token grants working Gmail access by
// reading the most recent message and checking it is non-empty.
func SmokeTest(ctx context.Context, cfg *oauth2.Config, refreshToken string) error {
	ts := cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})

	svc, err := gmailapi.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return fmt.Errorf("failed to create gmail service: %w", err)
	}

	msgs, err := svc.Users.Messages.List("me").MaxResults(1).Do()
	if err != nil {
		return fmt.Errorf("failed to list messages: %w", err)
	}
	if len(msgs.Messages) == 0 {
		return fmt.Errorf("mailbox appears empty — cannot verify access")
	}

	msg, err := svc.Users.Messages.Get("me", msgs.Messages[0].Id).Format("minimal").Do()
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}
	if msg.Id == "" {
		return fmt.Errorf("retrieved message has no ID")
	}

	return nil
}
