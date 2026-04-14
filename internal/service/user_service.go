package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/carlosEA28/ai-code-review/internal/domain"
	"github.com/carlosEA28/ai-code-review/internal/repository"
	"github.com/google/uuid"
)

type SaveGithubUserInput struct {
	GithubID    int
	GithubLogin string
	Name        string
	Email       string
	AvatarURL   string
	GithubToken string
}

type UserUseCase interface {
	UpsertGithubUser(ctx context.Context, input SaveGithubUserInput) (*domain.User, error)
}

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) UpsertGithubUser(ctx context.Context, input SaveGithubUserInput) (*domain.User, error) {
	if input.GithubID <= 0 {
		return nil, errors.New("github id is required")
	}

	login := strings.TrimSpace(input.GithubLogin)
	if login == "" {
		return nil, errors.New("github login is required")
	}

	user := domain.NewUser(
		uuid.New(),
		input.GithubID,
		login,
		strings.TrimSpace(input.Name),
		strings.TrimSpace(input.Email),
		strings.TrimSpace(input.AvatarURL),
		strings.TrimSpace(input.GithubToken),
	)

	persistedUser, err := s.userRepo.UpsertGithubUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("save github user: %w", err)
	}

	return persistedUser, nil
}
