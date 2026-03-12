package commands

import (
	"time"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpsertUser creates or updates a user record (keyed by Telegram ID).
func UpsertUser(db *gorm.DB, telegramID int64) error {
	user := entities.User{
		ID:           telegramID,
		LastActiveAt: time.Now(),
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_active_at", "updated_at"}),
	}).Create(&user).Error
}

func SetRegistered(db *gorm.DB, telegramID int64, registered bool) error {
	return db.Model(&entities.User{}).
		Where("id = ?", telegramID).
		Update("is_registered", registered).Error
}

func SetActive(db *gorm.DB, telegramID int64, active bool) error {
	return db.Model(&entities.User{}).
		Where("id = ?", telegramID).
		Update("is_active", active).Error
}

func SetWatching(db *gorm.DB, telegramID int64, chatID int64, watching bool) error {
	return db.Model(&entities.User{}).
		Where("id = ?", telegramID).
		Updates(map[string]interface{}{
			"is_watching":  watching,
			"watch_chat_id": chatID,
		}).Error
}

func UpdateLastChecked(db *gorm.DB, telegramID int64, t time.Time) error {
	return db.Model(&entities.User{}).
		Where("id = ?", telegramID).
		Update("last_checked_at", t).Error
}

func SetGmailAccountIndex(db *gorm.DB, telegramID int64, index int) error {
	return db.Model(&entities.User{}).
		Where("id = ?", telegramID).
		Update("gmail_account_index", index).Error
}
