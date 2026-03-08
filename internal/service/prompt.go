package service

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/commands"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/entities"
	"github.com/selfdeceited/tg-gmail-parser-bot/internal/db/queries"
)

// PromptService owns the business logic for managing summarization prompts.
type PromptService interface {
	ListPrompts(ctx context.Context, userID int64) ([]entities.Prompt, error)
	GetPrompt(ctx context.Context, id uuid.UUID) (*entities.Prompt, error)
	AddPrompt(ctx context.Context, userID int64, prompt, filter string) error
	UpdatePrompt(ctx context.Context, id uuid.UUID, prompt, filter string) error
	DeletePrompt(ctx context.Context, id uuid.UUID) error
}

type promptService struct {
	db *gorm.DB
}

// NewPromptService returns a GORM-backed PromptService.
func NewPromptService(db *gorm.DB) PromptService {
	return &promptService{db: db}
}

func (s *promptService) ListPrompts(_ context.Context, userID int64) ([]entities.Prompt, error) {
	return queries.GetActivePrompts(s.db, userID)
}

func (s *promptService) GetPrompt(_ context.Context, id uuid.UUID) (*entities.Prompt, error) {
	return queries.GetPromptByID(s.db, id)
}

func (s *promptService) AddPrompt(_ context.Context, userID int64, prompt, filter string) error {
	return commands.AddPrompt(s.db, userID, prompt, filter)
}

func (s *promptService) UpdatePrompt(_ context.Context, id uuid.UUID, prompt, filter string) error {
	return commands.UpdatePrompt(s.db, id, prompt, filter)
}

func (s *promptService) DeletePrompt(_ context.Context, id uuid.UUID) error {
	return commands.DeletePrompt(s.db, id)
}
