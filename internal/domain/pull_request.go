package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	PullRequestStatusMERGED PullRequestStatus = "MERGED"
	PullRequestStatusOPEN   PullRequestStatus = "OPEN"
)

type PullRequestStatus string

type PullRequest struct {
	// AssignedReviewers user_id назначенных ревьюверов (0..2)
	AssignedReviewers []string          `json:"assigned_reviewers"`
	AuthorId          uuid.UUID            `json:"author_id"`
	CreatedAt         *time.Time        `json:"createdAt"`
	MergedAt          *time.Time        `json:"mergedAt"`
	PullRequestId     string            `json:"pull_request_id"`
	PullRequestName   string            `json:"pull_request_name"`
	Status            PullRequestStatus `json:"status"`
}
