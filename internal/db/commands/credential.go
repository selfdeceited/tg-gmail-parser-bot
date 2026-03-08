package commands

import (
	"encoding/json"
	"fmt"

	internaldb "github.com/selfdeceited/tg-gmail-parser-bot/internal/db"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type credentialsPayload struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

// UpsertCredentials encrypts and stores (or replaces) the full OAuth credential set for a user.
func UpsertCredentials(db *gorm.DB, userID int64, clientID, clientSecret, refreshToken string) error {
	payload := credentialsPayload{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	encrypted, err := internaldb.EncryptToken(string(data))
	if err != nil {
		return err
	}

	cred := entities.Credential{
		UserID:               userID,
		EncryptedCredentials: encrypted,
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"encrypted_credentials", "updated_at"}),
	}).Create(&cred).Error
}

// DeleteCredential hard-deletes a user's credential.
func DeleteCredential(db *gorm.DB, userID int64) error {
	return db.Unscoped().Where("user_id = ?", userID).Delete(&entities.Credential{}).Error
}
