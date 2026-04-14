package domain

import (
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	ID            uuid.UUID
	UserId        uuid.UUID
	GithubRepoId  int
	WebhookId     int
	WebhookSecret string
	AutoReview    bool
	IsActive      bool
	Owner         string
	Name          string
	FullName      string
	DefaultBranch string
	BranchFilter  string
	LastRevieweAt time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PromptConfigs []PromptConfig
	PullRequests  []PullRequest
	User          User
}

func NewRepository(user User) *Repository {
	return &Repository{
		ID:            uuid.New(),
		UserId:        user.ID,
		User:          user,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		PromptConfigs: []PromptConfig{},
		PullRequests:  []PullRequest{},
	}
}
