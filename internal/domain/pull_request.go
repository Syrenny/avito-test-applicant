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
	PullRequestId   uuid.UUID         `json:"pull_request_id"`
	AuthorId        uuid.UUID         `json:"author_id"`
	CreatedAt       *time.Time        `json:"createdAt"`
	MergedAt        *time.Time        `json:"mergedAt"`
	PullRequestName string            `json:"pull_request_name"`
	Status          PullRequestStatus `json:"status"`
}

type PullRequestReviewers struct {
	PullRequestId uuid.UUID `json:"pull_request_id"`
	// AssignedReviewers user_id назначенных ревьюверов (0..2)
	AssignedReviewers []uuid.UUID `json:"assigned_reviewers"`
}

type PullRequestWithReviewers struct {
	PullRequest
	Reviewers []uuid.UUID `json:"reviewers"`
}

type PullRequestShort struct {
	PullRequestId   uuid.UUID         `json:"pull_request_id"`
	AuthorId        uuid.UUID         `json:"author_id"`
	PullRequestName string            `json:"pull_request_name"`
	Status          PullRequestStatus `json:"status"`
}
