package commands

import (
	"github.com/google/uuid"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"gorm.io/gorm"
)

func AddPrompt(db *gorm.DB, userID int64, prompt, filter string) error {
	p := entities.Prompt{
		UserID: userID,
		Prompt: prompt,
		Filter: filter,
	}
	return db.Create(&p).Error
}

func DeletePrompt(db *gorm.DB, id uuid.UUID) error {
	return db.Delete(&entities.Prompt{}, id).Error
}
