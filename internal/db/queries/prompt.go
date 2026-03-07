package queries

import (
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"gorm.io/gorm"
)

func GetActivePrompts(db *gorm.DB, userID int64) ([]entities.Prompt, error) {
	var prompts []entities.Prompt
	err := db.Where("user_id = ?", userID).Find(&prompts).Error
	return prompts, err
}
