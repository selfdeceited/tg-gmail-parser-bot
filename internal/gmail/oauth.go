package gmail

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// credentialsFile mirrors the structure of a GCP Desktop app credentials.json.
type credentialsFile struct {
	Installed struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"installed"`
}

var gmailScopes = []string{
	"https://www.googleapis.com/auth/gmail.readonly",
}

// ParseCredentials parses a GCP Desktop app credentials.json and returns an oauth2.Config.
func ParseCredentials(jsonBytes []byte) (*oauth2.Config, error) {
	var cf credentialsFile
	if err := json.Unmarshal(jsonBytes, &cf); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if cf.Installed.ClientID == "" || cf.Installed.ClientSecret == "" {
		return nil, fmt.Errorf("missing client_id or client_secret in credentials")
	}

	cfg := &oauth2.Config{
		ClientID:     cf.Installed.ClientID,
		ClientSecret: cf.Installed.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       gmailScopes,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}
	return cfg, nil
}

// BuildAuthURL returns the URL the user must visit to authorize Gmail access.
func BuildAuthURL(cfg *oauth2.Config) string {
	return cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCode exchanges the authorization code for a refresh token.
func ExchangeCode(ctx context.Context, cfg *oauth2.Config, code string) (string, error) {
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("token exchange failed: %w", err)
	}
	if token.RefreshToken == "" {
		return "", fmt.Errorf("no refresh token returned — ensure ApprovalForce and AccessTypeOffline were used")
	}
	return token.RefreshToken, nil
}
