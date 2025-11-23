package repo

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo/pgdb"
	"avito-test-applicant/pkg/postgres"
	"context"

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
	GetTeamById(
		ctx context.Context,
		teamId uuid.UUID,
	) (domain.Team, error)
}

type User interface {
	CreateUser(
		ctx context.Context,
		userId *uuid.UUID,
		username string,
		isActive bool,
		teamId uuid.UUID,
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
	GetUsersByTeam(
		ctx context.Context,
		teamId uuid.UUID,
	) ([]domain.User, error)
	UpdateUser(
		ctx context.Context,
		user domain.User,
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
	GetPullRequestsByIds(
		ctx context.Context,
		pullRequestIds []uuid.UUID,
	) ([]domain.PullRequest, error)
	SetMerged(
		ctx context.Context,
		pullRequestId uuid.UUID,
	) (domain.PullRequest, error)
}

type Reviewer interface {
	AssignOne(
		ctx context.Context,
		pullRequestId uuid.UUID,
		userId uuid.UUID,
	) error
	RemoveOne(
		ctx context.Context,
		pullRequestId uuid.UUID,
		userId uuid.UUID,
	) error
	ListReviewers(
		ctx context.Context,
		pullRequestId uuid.UUID,
	) ([]uuid.UUID, error)
	ListByUserId(
		ctx context.Context,
		userId uuid.UUID,
	) ([]uuid.UUID, error)
}

type Repositories struct {
	Team
	User
	PullRequest
	Reviewer
}

func NewRepositories(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		Team:        pgdb.NewTeamRepo(pg),
		User:        pgdb.NewUserRepo(pg),
		PullRequest: pgdb.NewPullRequestRepo(pg),
		Reviewer:    pgdb.NewReviewerRepo(pg),
	}
}
