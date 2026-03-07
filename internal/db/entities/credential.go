package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Credential struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID                int64     `gorm:"uniqueIndex;not null"`
	EncryptedRefreshToken string    `gorm:"type:text;not null"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             gorm.DeletedAt `gorm:"index"`
}

func (c *Credential) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
