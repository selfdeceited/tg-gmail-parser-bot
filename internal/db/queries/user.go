package queries

import (
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"gorm.io/gorm"
)

func GetUser(db *gorm.DB, telegramID int64) (*entities.User, error) {
	var user entities.User
	err := db.First(&user, telegramID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetWatchingUsers(db *gorm.DB) ([]entities.User, error) {
	var users []entities.User
	err := db.Where("is_watching = true").Find(&users).Error
	return users, err
}
