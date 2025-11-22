package pgdb

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/migrations/postgres"
	"context"

	"github.com/google/uuid"
)

type UserRepo struct {
	*postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) CreateUser(
	ctx context.Context,
	username string,
	team_name string,
	) (domain.User, error) {

	}

func (r *UserRepo) GetUserById(
	ctx context.Context,
	user_id uuid.UUID,
	) (domain.User, error) {

	}

func (r *UserRepo) SetIsActive(
	ctx context.Context,
	user_id uuid.UUID,
	is_active bool,
	) (domain.User, error) {

	}
