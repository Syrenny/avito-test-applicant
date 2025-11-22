package pgdb

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/migrations/postgres"
	"context"

	"github.com/google/uuid"
)


type TeamRepo struct {
	*postgres.Postgres
}

func NewTeamRepo(pg *postgres.Postgres) *TeamRepo {
	return &TeamRepo{pg}
}

func (r *TeamRepo) CreateTeam(
	ctx context.Context,
	team_name string,
) (uuid.UUID, error) {

}

func (r *TeamRepo) GetTeamByName(
	ctx context.Context,
	team_name string,
) (domain.Team, error)
