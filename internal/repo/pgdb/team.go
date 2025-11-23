package pgdb

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo/repoerrors"
	"avito-test-applicant/pkg/postgres"
	"context"
	"errors"

	"fmt"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TeamRepo struct {
	*postgres.Postgres
	getter *trmpgx.CtxGetter
}

func NewTeamRepo(pg *postgres.Postgres, getter *trmpgx.CtxGetter) *TeamRepo {
	return &TeamRepo{
		Postgres: pg,
		getter:   getter,
	}
}

func (r *TeamRepo) CreateTeam(
	ctx context.Context,
	teamId uuid.UUID,
	teamName string,
) (domain.Team, error) {
	query, args, err := r.Builder.
		Insert("teams").
		Columns("id", "team_name").
		Values(teamId, teamName).
		Suffix("RETURNING id, team_name").
		ToSql()
	if err != nil {
		return domain.Team{}, err
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.Pool)

	var t domain.Team
	err = conn.QueryRow(ctx, query, args...).Scan(
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

	conn := r.getter.DefaultTrOrDB(ctx, r.Pool)

	var t domain.Team
	err = conn.QueryRow(ctx, query, args...).Scan(
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

func (r *TeamRepo) GetTeamById(
	ctx context.Context,
	teamId uuid.UUID,
) (domain.Team, error) {
	query, args, err := r.Builder.
		Select("id", "team_name").
		From("teams").
		Where(squirrel.Eq{"id": teamId}).
		Limit(1).
		ToSql()
	if err != nil {
		return domain.Team{}, fmt.Errorf("build select team sql: %w", err)
	}

	conn := r.getter.DefaultTrOrDB(ctx, r.Pool)

	var t domain.Team
	err = conn.QueryRow(ctx, query, args...).Scan(
		&t.TeamId,
		&t.TeamName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Team{}, repoerrors.ErrNotFound
		}
		return domain.Team{}, fmt.Errorf("query team by id: %w", err)
	}

	return t, nil
}
