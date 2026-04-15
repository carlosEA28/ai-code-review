package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrRepositoryNotFound = errors.New("repository not found")

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
	Model              string
}

type WebhookRepository interface {
	UpsertPullRequestAndQueueReviewJob(ctx context.Context, input QueuePullRequestReviewInput) error
}

type webhookRepository struct {
	db *sql.DB
}

func NewWebhookRepository(db *sql.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) UpsertPullRequestAndQueueReviewJob(ctx context.Context, input QueuePullRequestReviewInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	const findRepositoryIDQuery = `
		SELECT id
		FROM repositories
		WHERE full_name = $1
		  AND is_active = true
		LIMIT 1
	`

	var repositoryID uuid.UUID
	err = tx.QueryRowContext(ctx, findRepositoryIDQuery, input.RepositoryFullName).Scan(&repositoryID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrRepositoryNotFound
	}
	if err != nil {
		return fmt.Errorf("find repository by full name: %w", err)
	}

	const upsertPullRequestQuery = `
		INSERT INTO pull_requests (
			repository_id,
			github_pr_id,
			number,
			title,
			body,
			author_login,
			base_branch,
			head_branch,
			head_sha,
			pr_url,
			diff_url
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		ON CONFLICT (repository_id, github_pr_id) DO UPDATE SET
			number = EXCLUDED.number,
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			author_login = EXCLUDED.author_login,
			base_branch = EXCLUDED.base_branch,
			head_branch = EXCLUDED.head_branch,
			head_sha = EXCLUDED.head_sha,
			pr_url = EXCLUDED.pr_url,
			diff_url = EXCLUDED.diff_url
		RETURNING id
	`

	var pullRequestID uuid.UUID
	err = tx.QueryRowContext(
		ctx,
		upsertPullRequestQuery,
		repositoryID,
		input.GithubPRID,
		input.Number,
		input.Title,
		input.Body,
		input.AuthorLogin,
		input.BaseBranch,
		input.HeadBranch,
		input.HeadSHA,
		input.PRURL,
		input.DiffURL,
	).Scan(&pullRequestID)
	if err != nil {
		return fmt.Errorf("upsert pull request: %w", err)
	}

	const insertReviewJobQuery = `
		INSERT INTO review_jobs (
			pull_request_id,
			status,
			model
		) VALUES (
			$1, 'queued', $2
		)
	`

	if _, err = tx.ExecContext(ctx, insertReviewJobQuery, pullRequestID, input.Model); err != nil {
		return fmt.Errorf("insert review job: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
