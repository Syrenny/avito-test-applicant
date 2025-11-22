package pgdb

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo/repoerrors"
	"avito-test-applicant/pkg/postgres"
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	*postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) CreateUser(
	ctx context.Context,
	userId *uuid.UUID,
	username string,
	isActive bool,
	teamId uuid.UUID,
) (domain.User, error) {
	var id uuid.UUID
	if userId != nil {
		id = *userId
	} else {
		id = uuid.New()
	}
	sql, args, err := r.Builder.
		Insert("users").
		Columns("id", "username", "is_active", "team_id").
		Values(id, username, isActive, teamId).
		Suffix("RETURNING id, username, team_id, is_active").
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("build insert user sql: %w", err)
	}

	var u domain.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.UserId,
		&u.Username,
		&u.TeamId,
		&u.IsActive,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("exec insert user: %w", err)
	}

	return u, nil
}

func (r *UserRepo) GetUserById(
	ctx context.Context,
	userId uuid.UUID,
) (domain.User, error) {
	sql, args, err := r.Builder.
		Select("id", "username", "team_id", "is_active").
		From("users").
		Where(squirrel.Eq{"id": userId}).
		Limit(1).
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("build select user sql: %w", err)
	}

	var u domain.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.UserId,
		&u.Username,
		&u.TeamId,
		&u.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, repoerrors.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("query user by id: %w", err)
	}

	return u, nil
}

func (r *UserRepo) SetIsActive(
	ctx context.Context,
	userId uuid.UUID,
	isActive bool,
) (domain.User, error) {
	sql, args, err := r.Builder.
		Update("users").
		Set("is_active", isActive).
		Where(squirrel.Eq{"id": userId}).
		Suffix("RETURNING id, username, team_id, is_active").
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("build update user sql: %w", err)
	}

	var u domain.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.UserId,
		&u.Username,
		&u.TeamId,
		&u.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, repoerrors.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("exec update user: %w", err)
	}

	return u, nil
}

func (r *UserRepo) GetUsersByTeam(
	ctx context.Context,
	teamId uuid.UUID,
) ([]domain.User, error) {
	sql, args, err := r.Builder.
		Select("id", "username", "team_id", "is_active").
		From("users").
		Where(squirrel.Eq{"team_id": teamId}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select users by team sql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query users by team: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(
			&u.UserId,
			&u.Username,
			&u.TeamId,
			&u.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user rows: %w", err)
	}

	return users, nil
}

func (r *UserRepo) UpdateUser(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	sql, args, err := r.Builder.
		Update("users").
		Set("username", user.Username).
		Set("team_id", user.TeamId).
		Set("is_active", user.IsActive).
		Where(squirrel.Eq{"id": user.UserId}).
		Suffix("RETURNING id, username, team_id, is_active").
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("build update user sql: %w", err)
	}

	var u domain.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.UserId,
		&u.Username,
		&u.TeamId,
		&u.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.User{}, repoerrors.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("exec update user: %w", err)
	}

	return u, nil
}
