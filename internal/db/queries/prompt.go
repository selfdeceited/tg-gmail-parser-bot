package queries

import (
	"github.com/google/uuid"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"gorm.io/gorm"
)

func GetActivePrompts(db *gorm.DB, userID int64) ([]entities.Prompt, error) {
	var prompts []entities.Prompt
	err := db.Where("user_id = ?", userID).Find(&prompts).Error
	return prompts, err
}

func GetPromptByID(db *gorm.DB, id uuid.UUID) (*entities.Prompt, error) {
	var prompt entities.Prompt
	err := db.First(&prompt, "id = ?", id).Error
	return &prompt, err
}
