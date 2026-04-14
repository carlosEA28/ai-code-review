package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	GithubID     int
	Githublogin  string
	Name         string
	Email        string
	avatarUrl    string
	githubToken  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Repositories []Repository
}

func NewUser(id uuid.UUID, githubID int, githublogin, name, email, avatarUrl, githubToken string) *User {
	return &User{
		ID:           id,
		GithubID:     githubID,
		Githublogin:  githublogin,
		Name:         name,
		Email:        email,
		avatarUrl:    avatarUrl,
		githubToken:  githubToken,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Repositories: []Repository{},
	}
}
