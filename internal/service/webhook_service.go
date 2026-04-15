package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/carlosEA28/ai-code-review/internal/repository"
)

type QueuePullRequestReviewInput struct {
	RepositoryFullName string
	GithubPRID         int64
	Number             int
	Title              string
	Body               string
	AuthorLogin        string
	BaseBranch         string
	HeadBranch         string
	HeadSHA            string
	PRURL              string
	DiffURL            string
}

type WebhookUseCase interface {
	QueuePullRequestReview(ctx context.Context, input QueuePullRequestReviewInput) error
}

type WebhookService struct {
	webhookRepo  repository.WebhookRepository
	defaultModel string
}

func NewWebhookService(webhookRepo repository.WebhookRepository, defaultModel string) *WebhookService {
	model := strings.TrimSpace(defaultModel)
	if model == "" {
		model = "claude-sonnet-4-5"
	}

	return &WebhookService{
		webhookRepo:  webhookRepo,
		defaultModel: model,
	}
}

func (s *WebhookService) QueuePullRequestReview(ctx context.Context, input QueuePullRequestReviewInput) error {
	if strings.TrimSpace(input.RepositoryFullName) == "" {
		return errors.New("repository full name is required")
	}

	if input.GithubPRID <= 0 {
		return errors.New("github pr id is required")
	}

	if input.Number <= 0 {
		return errors.New("pull request number is required")
	}

	if strings.TrimSpace(input.AuthorLogin) == "" {
		return errors.New("author login is required")
	}

	if strings.TrimSpace(input.BaseBranch) == "" {
		return errors.New("base branch is required")
	}

	if strings.TrimSpace(input.HeadBranch) == "" {
		return errors.New("head branch is required")
	}

	if strings.TrimSpace(input.HeadSHA) == "" {
		return errors.New("head sha is required")
	}

	err := s.webhookRepo.UpsertPullRequestAndQueueReviewJob(ctx, repository.QueuePullRequestReviewInput{
		RepositoryFullName: strings.TrimSpace(input.RepositoryFullName),
		GithubPRID:         input.GithubPRID,
		Number:             input.Number,
		Title:              strings.TrimSpace(input.Title),
		Body:               strings.TrimSpace(input.Body),
		AuthorLogin:        strings.TrimSpace(input.AuthorLogin),
		BaseBranch:         strings.TrimSpace(input.BaseBranch),
		HeadBranch:         strings.TrimSpace(input.HeadBranch),
		HeadSHA:            strings.TrimSpace(input.HeadSHA),
		PRURL:              strings.TrimSpace(input.PRURL),
		DiffURL:            strings.TrimSpace(input.DiffURL),
		Model:              s.defaultModel,
	})
	if err != nil {
		if errors.Is(err, repository.ErrRepositoryNotFound) {
			return err
		}
		return fmt.Errorf("queue pull request review: %w", err)
	}

	return nil
}
