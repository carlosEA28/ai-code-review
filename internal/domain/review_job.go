package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReviewJobStatus string

const (
	ReviewJobStatusQueued     ReviewJobStatus = "queued"
	ReviewJobStatusProcessing ReviewJobStatus = "processing"
	ReviewJobStatusDone       ReviewJobStatus = "done"
	ReviewJobStatusFailed     ReviewJobStatus = "failed"
)

type ReviewJob struct {
	ID             uuid.UUID
	PullRequestId  uuid.UUID
	PromptConfigId uuid.UUID
	Status         ReviewJobStatus
	Model          string
	InputTokens    int
	OutputTokens   int
	DurationMs     int
	ErrorMessage   string
	RetryCount     int
	QueuedAt       time.Time
	StartedAt      time.Time
	FinishedAt     time.Time
	PullRequest    PullRequest
	PromptConfig   PromptConfig
	ReviewComments []ReviewComment
}

func NewReviewJob(id, pullRequestId uuid.UUID, promptConfigId uuid.UUID, status ReviewJobStatus, model string) *ReviewJob {
	if status == "" {
		status = ReviewJobStatusQueued
	}

	return &ReviewJob{
		ID:             id,
		PullRequestId:  pullRequestId,
		PromptConfigId: promptConfigId,
		Status:         status,
		Model:          model,
		RetryCount:     0,
		QueuedAt:       time.Now(),
		ReviewComments: []ReviewComment{},
	}
}
