package commands

import (
	internaldb "github.com/selfdeceited/tg-gmail-parser-bot/internal/db"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpsertRefreshToken encrypts and stores (or replaces) the OAuth refresh token for a user.
func UpsertRefreshToken(db *gorm.DB, userID int64, refreshToken string) error {
	encrypted, err := internaldb.EncryptToken(refreshToken)
	if err != nil {
		return err
	}

	cred := entities.Credential{
		UserID:                userID,
		EncryptedRefreshToken: encrypted,
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"encrypted_refresh_token", "updated_at"}),
	}).Create(&cred).Error
}

// DeleteCredential hard-deletes a user's credential (used by the cleanup job).
func DeleteCredential(db *gorm.DB, userID int64) error {
	return db.Unscoped().Where("user_id = ?", userID).Delete(&entities.Credential{}).Error
}
