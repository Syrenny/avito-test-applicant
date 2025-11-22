package service

import (
	"context"

	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"avito-test-applicant/pkg/postgres"

	"github.com/google/uuid"
)

type Team interface {
	CreateTeam(
		ctx context.Context,
		teamName string,
	) (domain.Team, error)
	GetTeamByName(
		ctx context.Context,
		teamName string,
	) (domain.Team, error)
}

type User interface {
	CreateUser(
		ctx context.Context,
		username string,
		teamName string,
	) (domain.User, error)
	GetUserById(
		ctx context.Context,
		userId uuid.UUID,
	) (domain.User, error)
	SetIsActive(
		ctx context.Context,
		userId uuid.UUID,
		isActive bool,
	) (domain.User, error)
}

type PullRequest interface {
	CreatePullRequest(
		ctx context.Context,
		pullRequestId uuid.UUID,
		pullRequestName string,
		authorId uuid.UUID,
	) (domain.PullRequest, error)
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
	) (domain.PullRequest, error)
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
		Team:        NewTeamService(deps.Repos),
		User:        NewUserService(deps.Repos, deps.TrManager),
		PullRequest: NewPullRequestService(deps.Repos, deps.TrManager),
	}
}
