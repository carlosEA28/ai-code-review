package domain

import (
	"time"

	"github.com/google/uuid"
)

type CommentSeverity string

const (
	CommentSeverityCritical   CommentSeverity = "critical"
	CommentSeverityWarning    CommentSeverity = "warning"
	CommentSeveritySuggestion CommentSeverity = "suggestion"
)

type ReviewComment struct {
	ID              uuid.UUID
	ReviewJobId     uuid.UUID
	FilePath        string
	LineNumber      int
	DiffPosition    int
	Severity        CommentSeverity
	Body            string
	CodeSnippet     string
	GithubCommentId int
	PostedToGithub  bool
	PostedAt        time.Time
	CreatedAt       time.Time
	ReviewJob       ReviewJob
}

func NewReviewComment(
	id, reviewJobId uuid.UUID,
	filePath string,
	lineNumber, diffPosition int,
	severity CommentSeverity,
	body, codeSnippet string,
) *ReviewComment {
	return &ReviewComment{
		ID:           id,
		ReviewJobId:  reviewJobId,
		FilePath:     filePath,
		LineNumber:   lineNumber,
		DiffPosition: diffPosition,
		Severity:     severity,
		Body:         body,
		CodeSnippet:  codeSnippet,
		CreatedAt:    time.Now(),
	}
}
