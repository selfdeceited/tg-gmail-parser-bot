package entities

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                 int64          `gorm:"primaryKey"`
	IsRegistered       bool
	IsActive           bool
	IsWatching         bool
	WatchChatID        int64
	GmailAccountIndex  int
	LastCheckedAt      *time.Time
	LastActiveAt       time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}
