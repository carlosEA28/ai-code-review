package domain

import (
	"time"

	"github.com/google/uuid"
)

type PullRequest struct {
	ID           uuid.UUID
	RepositoryId uuid.UUID
	GithubPrId   int
	Number       int
	Title        string
	Body         string
	AuthorLogin  string
	BaseBranch   string
	HeadBranch   string
	HeadSha      string
	PrUrl        string
	DiffUrl      string
	CreatedAt    time.Time
	Repository   Repository
	ReviewJobs   []ReviewJob
}

func NewPullRequest(
	id, repositoryId uuid.UUID,
	githubPrId, number int,
	title, body, authorLogin, baseBranch, headBranch, headSha, prUrl, diffUrl string,
) *PullRequest {
	return &PullRequest{
		ID:           id,
		RepositoryId: repositoryId,
		GithubPrId:   githubPrId,
		Number:       number,
		Title:        title,
		Body:         body,
		AuthorLogin:  authorLogin,
		BaseBranch:   baseBranch,
		HeadBranch:   headBranch,
		HeadSha:      headSha,
		PrUrl:        prUrl,
		DiffUrl:      diffUrl,
		CreatedAt:    time.Now(),
		ReviewJobs:   []ReviewJob{},
	}
}
