package pgdb

import (
	"avito-test-applicant/internal/repo/repoerrors"
	"avito-test-applicant/pkg/postgres"
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type ReviewerRepo struct {
	*postgres.Postgres
}

func NewReviewerRepo(pg *postgres.Postgres) *ReviewerRepo {
	return &ReviewerRepo{pg}
}

func (r *ReviewerRepo) AssignOne(
	ctx context.Context,
	pullRequestId uuid.UUID,
	userId uuid.UUID,
) error {
	sql, args, _ := r.Builder.
		Insert("pr_reviewers").
		Columns("pr_id", "user_id").
		Values(pullRequestId, userId).
		ToSql()

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return repoerrors.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *ReviewerRepo) RemoveOne(
	ctx context.Context,
	pullRequestId uuid.UUID,
	userId uuid.UUID,
) error {
	sql, args, err := r.Builder.
		Delete("pr_reviewers").
		Where(squirrel.Eq{
			"pr_id":   pullRequestId,
			"user_id": userId,
		}).
		ToSql()
	if err != nil {
		return err
	}

	cmdTag, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return repoerrors.ErrNotFound
	}

	return nil
}

func (r *ReviewerRepo) ListReviewers(
	ctx context.Context,
	pullRequestId uuid.UUID,
) ([]uuid.UUID, error) {
	sql, args, err := r.Builder.
		Select("user_id").
		From("pr_reviewers").
		Where(squirrel.Eq{
			"pr_id": pullRequestId,
		}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []uuid.UUID
	for rows.Next() {
		var userId uuid.UUID
		if err := rows.Scan(&userId); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, userId)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviewers, nil
}

func (r *ReviewerRepo) ListByUserId(
	ctx context.Context,
	userId uuid.UUID,
) ([]uuid.UUID, error) {
	sql, args, err := r.Builder.
		Select("pr_id").
		From("pr_reviewers").
		Where(squirrel.Eq{"user_id": userId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		prIDs = append(prIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prIDs, nil
}
