package pgdb

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/migrations/postgres"
	"context"

	"github.com/google/uuid"
)


type PullRequestRepo struct {
	*postgres.Postgres
}

func NewPullRequestRepo(pg *postgres.Postgres) *PullRequestRepo {
	return &PullRequestRepo{pg}
}

func (r *PullRequestRepo) CreatePullRequest(
	ctx context.Context,
	pull_request_id uuid.UUID,
	pull_request_name string,
	author_id uuid.UUID,
	) (domain.PullRequest, error) {

	}

func (r *PullRequestRepo) GetPullRequestById(
	ctx context.Context,
	pull_request_id uuid.UUID,
	) (domain.PullRequest, error) {

	}

func (r *PullRequestRepo) SetMerged(
	ctx context.Context,
	pull_request_id uuid.UUID,
	) (domain.PullRequest, error) {

	}

func (r *PullRequestRepo) Reassign(
	ctx context.Context,
	pull_request_id uuid.UUID,
	old_user_id uuid.UUID,
	) (domain.PullRequest, error) {

	}
