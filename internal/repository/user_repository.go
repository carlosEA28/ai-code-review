package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/carlosEA28/ai-code-review/internal/domain"
)

type UserRepository interface {
	UpsertGithubUser(ctx context.Context, user *domain.User) (*domain.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) UpsertGithubUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	const query = `
		INSERT INTO users (
			id, github_id, github_login, name, email, avatar_url, github_token
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (github_id) DO UPDATE SET
			github_login = EXCLUDED.github_login,
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			avatar_url = EXCLUDED.avatar_url,
			github_token = EXCLUDED.github_token,
			updated_at = now()
		RETURNING
			id,
			github_id,
			github_login,
			COALESCE(name, ''),
			COALESCE(email::text, ''),
			COALESCE(avatar_url, ''),
			COALESCE(github_token, ''),
			created_at,
			updated_at
	`

	persistedUser := &domain.User{}

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.GithubID,
		user.GithubLogin,
		user.Name,
		user.Email,
		user.AvatarUrl,
		user.GithubToken,
	).Scan(
		&persistedUser.ID,
		&persistedUser.GithubID,
		&persistedUser.GithubLogin,
		&persistedUser.Name,
		&persistedUser.Email,
		&persistedUser.AvatarUrl,
		&persistedUser.GithubToken,
		&persistedUser.CreatedAt,
		&persistedUser.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert github user: %w", err)
	}

	return persistedUser, nil
}
