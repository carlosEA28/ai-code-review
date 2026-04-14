package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	GithubID     int
	GithubLogin  string
	Name         string
	Email        string
	AvatarUrl    string
	GithubToken  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Repositories []Repository
}

func NewUser(id uuid.UUID, githubID int, githubLogin, name, email, avatarUrl, githubToken string) *User {
	return &User{
		ID:           id,
		GithubID:     githubID,
		GithubLogin:  githubLogin,
		Name:         name,
		Email:        email,
		AvatarUrl:    avatarUrl,
		GithubToken:  githubToken,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Repositories: []Repository{},
	}
}
