package queries

import (
	"encoding/json"
	"fmt"

	internaldb "github.com/selfdeceited/tg-gmail-parser-bot/internal/db"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"gorm.io/gorm"
)

type StoredCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

// ListAllCredentials returns all raw credential rows (encrypted). Used for key rotation.
func ListAllCredentials(db *gorm.DB) ([]entities.Credential, error) {
	var creds []entities.Credential
	return creds, db.Find(&creds).Error
}

// GetCredentials decrypts and returns the stored OAuth credentials for a user.
func GetCredentials(db *gorm.DB, userID int64) (*StoredCredentials, error) {
	var cred entities.Credential
	if err := db.Where("user_id = ?", userID).First(&cred).Error; err != nil {
		return nil, err
	}
	plaintext, err := internaldb.DecryptCredentials(cred.EncryptedCredentials, cred.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}
	var stored StoredCredentials
	if err := json.Unmarshal([]byte(plaintext), &stored); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}
	return &stored, nil
}
