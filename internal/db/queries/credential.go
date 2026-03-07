package queries

import (
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	internaldb "github.com/selfdeceited/tg-gmail-parser-bot/internal/db"
	"gorm.io/gorm"
)

// GetRefreshToken returns the decrypted OAuth refresh token for the user.
func GetRefreshToken(db *gorm.DB, userID int64) (string, error) {
	var cred entities.Credential
	err := db.Where("user_id = ?", userID).First(&cred).Error
	if err != nil {
		return "", err
	}
	return internaldb.DecryptToken(cred.EncryptedRefreshToken)
}
