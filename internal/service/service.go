package service

import (
	"context"

	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"avito-test-applicant/pkg/postgres"

	"github.com/google/uuid"
)

type Team interface {
	CreateTeamWithUsers(
		ctx context.Context,
		teamName string,
		members []domain.UserInput,
	) (domain.TeamWithUsers, error)
	GetTeamByName(
		ctx context.Context,
		teamName string,
	) (domain.TeamWithUsers, error)
}

type User interface {
	GetUserById(
		ctx context.Context,
		userId uuid.UUID,
	) (domain.User, error)
	SetIsActive(
		ctx context.Context,
		userId uuid.UUID,
		isActive bool,
	) (domain.UserWithTeamName, error)
}

type PullRequest interface {
	CreateAndAssignPullRequest(
		ctx context.Context,
		pullRequestId uuid.UUID,
		pullRequestName string,
		authorId uuid.UUID,
	) (domain.PullRequestWithReviewers, error)
	GetPullRequestById(
		ctx context.Context,
		pullRequestId uuid.UUID,
	) (domain.PullRequest, error)
	SetMerged(
		ctx context.Context,
		pullRequestId uuid.UUID,
	) (domain.PullRequest, error)
	Reassign(
		ctx context.Context,
		pullRequestId uuid.UUID,
		oldUserId uuid.UUID,
	) (domain.PullRequestWithReviewers, error)
	GetAssignedReviewsByUserId(
		ctx context.Context,
		userId uuid.UUID,
	) ([]domain.PullRequestShort, error)
}

type Services struct {
	Team        Team
	User        User
	PullRequest PullRequest
}

type ServicesDependencies struct {
	Repos     *repo.Repositories
	TrManager *postgres.TransactionManager
}

func NewServices(deps ServicesDependencies) *Services {
	return &Services{
		Team:        NewTeamService(deps.Repos, deps.TrManager),
		User:        NewUserService(deps.Repos, deps.TrManager),
		PullRequest: NewPullRequestService(deps.Repos, deps.TrManager),
	}
}
