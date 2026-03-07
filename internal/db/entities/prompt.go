package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Prompt struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    int64     `gorm:"index;not null"`
	Prompt    string    `gorm:"type:text;not null"`
	Filter    string    `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (p *Prompt) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
