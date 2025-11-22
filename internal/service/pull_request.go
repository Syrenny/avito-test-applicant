package service

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"context"

	"github.com/google/uuid"
)

type PullRequestService struct {
	pullRequestRepo repo.PullRequest
}

func NewPullRequestService(repos *repo.Repositories) *PullRequestService {
	return &PullRequestService{
		pullRequestRepo: repos.PullRequest,
	}
}

func (s *PullRequestService) CreatePullRequest(
	ctx context.Context, pull_request_id uuid.UUID, pull_request_name string, author_id uuid.UUID,
) (domain.PullRequest, error) {

}

func (s *PullRequestService) GetPullRequestById(
	ctx context.Context, pull_request_id uuid.UUID,
) (domain.PullRequest, error) {
}

func (s *PullRequestService) SetMerged(
	ctx context.Context, pull_request_id uuid.UUID,
) (domain.PullRequest, error) {
}

func (s *PullRequestService) Reassign(
	ctx context.Context, pull_request_id uuid.UUID, old_user_id uuid.UUID,
) (domain.PullRequest, error) {
}
