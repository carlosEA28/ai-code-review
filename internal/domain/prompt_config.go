package domain

import (
	"time"

	"github.com/google/uuid"
)

type PromptConfig struct {
	ID           uuid.UUID
	RepositoryId uuid.UUID
	SystemPrompt string
	Model        string
	MaxTokens    int
	IsActive     bool
	CreatedAt    time.Time
	Repository   Repository
	ReviewJobs   []ReviewJob
}

func NewPromptConfig(id, repositoryId uuid.UUID, systemPrompt, model string, maxTokens int, isActive bool) *PromptConfig {
	return &PromptConfig{
		ID:           id,
		RepositoryId: repositoryId,
		SystemPrompt: systemPrompt,
		Model:        model,
		MaxTokens:    maxTokens,
		IsActive:     isActive,
		CreatedAt:    time.Now(),
		ReviewJobs:   []ReviewJob{},
	}
}
