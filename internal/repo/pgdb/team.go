package pgdb

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo/repoerrors"
	"avito-test-applicant/pkg/postgres"
	"context"
	"errors"

	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TeamRepo struct {
	*postgres.Postgres
}

func NewTeamRepo(pg *postgres.Postgres) *TeamRepo {
	return &TeamRepo{pg}
}

func (r *TeamRepo) CreateTeam(
	ctx context.Context,
	teamName string,
) (domain.Team, error) {
	query, args, err := r.Builder.
		Insert("teams").
		Columns("team_name").
		Values(teamName).
		Suffix("RETURNING id, team_name").
		ToSql()
	if err != nil {
		return domain.Team{}, err
	}

	var t domain.Team
	err = r.Pool.QueryRow(ctx, query, args...).Scan(
		&t.TeamId,
		&t.TeamName,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return domain.Team{}, repoerrors.ErrAlreadyExists
			}
		}
		return domain.Team{}, fmt.Errorf("exec insert team: %w", err)
	}

	return t, nil
}

func (r *TeamRepo) GetTeamByName(
	ctx context.Context,
	teamName string,
) (domain.Team, error) {
	query, args, err := r.Builder.
		Select("id", "team_name").
		From("teams").
		Where(squirrel.Eq{"team_name": teamName}).
		Limit(1).
		ToSql()
	if err != nil {
		return domain.Team{}, fmt.Errorf("build select team sql: %w", err)
	}

	var t domain.Team
	err = r.Pool.QueryRow(ctx, query, args...).Scan(
		&t.TeamId,
		&t.TeamName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Team{}, repoerrors.ErrNotFound
		}
		return domain.Team{}, fmt.Errorf("query team by name: %w", err)
	}

	return t, nil
}
