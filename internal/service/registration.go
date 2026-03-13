package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	internaldb "github.com/selfdeceited/tg-gmail-parser-bot/internal/db"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/commands"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/queries"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/gmail"
)

// ErrNotRegistered is returned when an operation requires a linked Gmail account
// but none exists for the user. Callers should check with errors.Is.
var ErrNotRegistered = errors.New("user not registered")

// RegistrationService owns the business logic for Gmail account linking.
type RegistrationService interface {
	// SaveCredentials encrypts and persists OAuth credentials and marks the user as registered.
	SaveCredentials(ctx context.Context, userID int64, clientID, clientSecret, refreshToken string) error
	// VerifyCredentials fetches stored credentials and runs a live Gmail smoke test.
	VerifyCredentials(ctx context.Context, userID int64) error
	// ClearCredentials deletes stored credentials and marks the user as unregistered.
	ClearCredentials(ctx context.Context, userID int64) error
	// IsRegistered returns true if the user has completed registration.
	IsRegistered(ctx context.Context, userID int64) (bool, error)
	// RotateCredentials re-encrypts all stored credentials with the current key version.
	// Call this after bumping TOKEN_ENCRYPTION_KEY_CURRENT to migrate existing rows.
	RotateCredentials(ctx context.Context) (rotated int, err error)
	// GetGmailAccountIndex returns the Gmail account index (the /u/<N>/ slot) for the user.
	GetGmailAccountIndex(ctx context.Context, userID int64) (int, error)
	// SetGmailAccountIndex updates the Gmail account index for the user.
	SetGmailAccountIndex(ctx context.Context, userID int64, index int) error
}

type registrationService struct {
	db        *gorm.DB
	ioTimeout time.Duration
}

// NewRegistrationService returns a GORM-backed RegistrationService.
func NewRegistrationService(db *gorm.DB, ioTimeout time.Duration) RegistrationService {
	return &registrationService{db: db, ioTimeout: ioTimeout}
}

func (s *registrationService) SaveCredentials(ctx context.Context, userID int64, clientID, clientSecret, refreshToken string) error {
	ioCtx, cancel := context.WithTimeout(ctx, s.ioTimeout)
	defer cancel()
	db := s.db.WithContext(ioCtx)
	if err := commands.UpsertUser(db, userID); err != nil {
		return err
	}
	if err := commands.UpsertCredentials(db, userID, clientID, clientSecret, refreshToken); err != nil {
		return err
	}
	return commands.SetRegistered(db, userID, true)
}

func (s *registrationService) VerifyCredentials(ctx context.Context, userID int64) error {
	ioCtx, cancel := context.WithTimeout(ctx, s.ioTimeout)
	defer cancel()
	db := s.db.WithContext(ioCtx)
	creds, err := queries.GetCredentials(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotRegistered
		}
		return err
	}
	return gmail.VerifyRefreshToken(ioCtx, creds.ClientID, creds.ClientSecret, creds.RefreshToken)
}

func (s *registrationService) IsRegistered(ctx context.Context, userID int64) (bool, error) {
	ioCtx, cancel := context.WithTimeout(ctx, s.ioTimeout)
	defer cancel()
	user, err := queries.GetUser(s.db.WithContext(ioCtx), userID)
	if err != nil {
		return false, nil // user not found → not registered
	}
	return user.IsRegistered, nil
}

func (s *registrationService) ClearCredentials(ctx context.Context, userID int64) error {
	ioCtx, cancel := context.WithTimeout(ctx, s.ioTimeout)
	defer cancel()
	db := s.db.WithContext(ioCtx)
	if err := commands.DeleteCredential(db, userID); err != nil {
		return err
	}
	return commands.SetRegistered(db, userID, false)
}

func (s *registrationService) GetGmailAccountIndex(ctx context.Context, userID int64) (int, error) {
	ioCtx, cancel := context.WithTimeout(ctx, s.ioTimeout)
	defer cancel()
	user, err := queries.GetUser(s.db.WithContext(ioCtx), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrNotRegistered
		}
		return 0, err
	}
	return user.GmailAccountIndex, nil
}

func (s *registrationService) SetGmailAccountIndex(ctx context.Context, userID int64, index int) error {
	ioCtx, cancel := context.WithTimeout(ctx, s.ioTimeout)
	defer cancel()
	return commands.SetGmailAccountIndex(s.db.WithContext(ioCtx), userID, index)
}

func (s *registrationService) RotateCredentials(ctx context.Context) (int, error) {
	ioCtx, cancel := context.WithTimeout(ctx, s.ioTimeout)
	defer cancel()
	db := s.db.WithContext(ioCtx)
	// todo: consider pagination on larger datasets
	rows, err := queries.ListAllCredentials(db)
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
		if err := db.Model(&cred).Update("encrypted_credentials", reencrypted).Error; err != nil {
			logrus.WithError(err).WithField("user_id", cred.UserID).Error("rotate: failed to save rotated credentials")
			continue
		}
		rotated++
		logrus.WithField("user_id", cred.UserID).Info("rotate: credentials re-encrypted with current key version")
	}
	return rotated, nil
}
