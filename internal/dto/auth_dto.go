package dto

import "time"

type GithubAuthResponseDto struct {
	ID          string `json:"id"`
	GithubID    int    `json:"github_id"`
	GithubLogin string `json:"github_login"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	Token       string `json:"token"`
	ExpiresAt   string `json:"expires_at"`
}

type MeResponseDto struct {
	ID          string    `json:"id"`
	GithubID    int       `json:"github_id"`
	GithubLogin string    `json:"github_login"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}
