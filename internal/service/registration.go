package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	internaldb "github.com/selfdeceited/tg-gmail-parser-bot/internal/db"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/commands"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/queries"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/gmail"
)

// RegistrationService owns the business logic for Gmail account linking.
type RegistrationService interface {
	// SaveCredentials encrypts and persists OAuth credentials and marks the user as registered.
	SaveCredentials(ctx context.Context, userID int64, clientID, clientSecret, refreshToken string) error
	// VerifyCredentials fetches stored credentials and runs a live Gmail smoke test.
	VerifyCredentials(ctx context.Context, userID int64) error
	// ClearCredentials deletes stored credentials and marks the user as unregistered.
	ClearCredentials(ctx context.Context, userID int64) error
	// RotateCredentials re-encrypts all stored credentials with the current key version.
	// Call this after bumping TOKEN_ENCRYPTION_KEY_CURRENT to migrate existing rows.
	RotateCredentials(ctx context.Context) (rotated int, err error)
}

type registrationService struct {
	db *gorm.DB
}

// NewRegistrationService returns a GORM-backed RegistrationService.
func NewRegistrationService(db *gorm.DB) RegistrationService {
	return &registrationService{db: db}
}

func (s *registrationService) SaveCredentials(ctx context.Context, userID int64, clientID, clientSecret, refreshToken string) error {
	if err := commands.UpsertCredentials(s.db, userID, clientID, clientSecret, refreshToken); err != nil {
		return err
	}
	return commands.SetRegistered(s.db, userID, true)
}

func (s *registrationService) VerifyCredentials(ctx context.Context, userID int64) error {
	creds, err := queries.GetCredentials(s.db, userID)
	if err != nil {
		return err
	}
	return gmail.VerifyRefreshToken(ctx, creds.ClientID, creds.ClientSecret, creds.RefreshToken)
}

func (s *registrationService) ClearCredentials(ctx context.Context, userID int64) error {
	if err := commands.DeleteCredential(s.db, userID); err != nil {
		return err
	}
	return commands.SetRegistered(s.db, userID, false)
}

func (s *registrationService) RotateCredentials(ctx context.Context) (int, error) {
	rows, err := queries.ListAllCredentials(s.db)
	if err != nil {
		return 0, fmt.Errorf("failed to list credentials: %w", err)
	}
	rotated := 0
	for _, cred := range rows {
		plaintext, err := internaldb.DecryptCredentials(cred.EncryptedCredentials, cred.UserID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", cred.UserID).Error("rotate: failed to decrypt credentials")
			continue
		}
		reencrypted, err := internaldb.EncryptCredentials(plaintext, cred.UserID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", cred.UserID).Error("rotate: failed to re-encrypt credentials")
			continue
		}
		if err := s.db.Model(&cred).Update("encrypted_credentials", reencrypted).Error; err != nil {
			logrus.WithError(err).WithField("user_id", cred.UserID).Error("rotate: failed to save rotated credentials")
			continue
		}
		rotated++
		logrus.WithField("user_id", cred.UserID).Info("rotate: credentials re-encrypted with current key version")
	}
	return rotated, nil
}
