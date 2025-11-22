package service

import (
	"context"

	"avito-test-applicant/internal/repo"

	"avito-test-applicant/internal/domain"

	"github.com/google/uuid"
)

type Team interface {
	CreateTeam(
		ctx context.Context,
		team_name string,
		) (domain.Team, error)
	GetTeamByName(
		ctx context.Context,
		team_name string,
		) (domain.Team, error)
}

type User interface {
	CreateUser(
		ctx context.Context,
		username string,
		team_name string,
		) (domain.User, error)
	GetUserById(
		ctx context.Context,
		user_id uuid.UUID,
		) (domain.User, error)
	SetIsActive(
		ctx context.Context,
		user_id uuid.UUID,
		is_active bool,
		) (domain.User, error)
}

type PullRequest interface {
	CreatePullRequest(
		ctx context.Context,
		pull_request_id uuid.UUID,
		pull_request_name string,
		author_id uuid.UUID,
		) (domain.PullRequest, error)
	GetPullRequestById(
		ctx context.Context,
		pull_request_id uuid.UUID,
		) (domain.PullRequest, error)
	SetMerged(
		ctx context.Context,
		pull_request_id uuid.UUID,
		) (domain.PullRequest, error)
	Reassign(
		ctx context.Context,
		pull_request_id uuid.UUID,
		old_user_id uuid.UUID,
		) (domain.PullRequest, error)
}

type Services struct {
	Team    Team
	User    User
	PullRequest PullRequest
}

type ServicesDependencies struct {
	Repos *repo.Repositories
}

func NewServices(deps ServicesDependencies) *Services {
	return &Services{
		Team:    NewTeamService(deps.Repos),
		User:    NewUserService(deps.Repos),
		PullRequest: NewPullRequestService(deps.Repos),
	}
}
