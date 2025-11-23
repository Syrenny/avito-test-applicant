package pgdb

import (
	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo/repoerrors"
	"avito-test-applicant/pkg/postgres"
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PullRequestRepo struct {
	*postgres.Postgres
}

func NewPullRequestRepo(pg *postgres.Postgres) *PullRequestRepo {
	return &PullRequestRepo{pg}
}

func toDomainPullRequestStatus(smallint int) domain.PullRequestStatus {
	if smallint == 1 {
		return domain.PullRequestStatusMERGED
	}
	return domain.PullRequestStatusOPEN
}

func (r *PullRequestRepo) CreatePullRequest(
	ctx context.Context,
	pullRequestId uuid.UUID,
	pullRequestName string,
	authorId uuid.UUID,
) (domain.PullRequest, error) {
	sql, args, err := r.Builder.
		Insert("pull_requests").
		Columns("id", "pr_name", "author_id", "pr_status", "created_at").
		Values(pullRequestId, pullRequestName, authorId, 0, time.Now().UTC()).
		Suffix("RETURNING id, pr_name, author_id, pr_status, created_at, merged_at").
		ToSql()
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("build insert PR sql: %w", err)
	}

	var pr domain.PullRequest
	var statusSmallint int
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&statusSmallint,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return domain.PullRequest{}, repoerrors.ErrAlreadyExists
		}
		return domain.PullRequest{}, fmt.Errorf("exec insert PR: %w", err)

	}

	pr.Status = toDomainPullRequestStatus(statusSmallint)

	return pr, nil
}

func (r *PullRequestRepo) GetPullRequestById(
	ctx context.Context,
	pullRequestId uuid.UUID,
) (domain.PullRequest, error) {
	sql, args, err := r.Builder.
		Select("id", "pr_name", "author_id", "pr_status", "created_at", "merged_at").
		From("pull_requests").
		Where(squirrel.Eq{"id": pullRequestId}).
		Limit(1).
		ToSql()
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("build select PR sql: %w", err)
	}

	var pr domain.PullRequest
	var statusSmallint int
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&statusSmallint,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.PullRequest{}, repoerrors.ErrNotFound
		}
		return domain.PullRequest{}, fmt.Errorf("query PR by id: %w", err)
	}

	pr.Status = toDomainPullRequestStatus(statusSmallint)

	return pr, nil
}

func (r *PullRequestRepo) GetPullRequestsByIds(
	ctx context.Context,
	ids []uuid.UUID,
) ([]domain.PullRequest, error) {
	if len(ids) == 0 {
		return []domain.PullRequest{}, nil
	}

	sql, args, err := r.Builder.
		Select("id", "pr_name", "author_id", "pr_status", "created_at", "merged_at").
		From("pull_requests").
		Where(squirrel.Eq{"id": ids}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select PRs sql: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query PRs by ids: %w", err)
	}
	defer rows.Close()

	// collect found PRs in a map for ordering later
	found := make(map[uuid.UUID]domain.PullRequest, len(ids))
	for rows.Next() {
		var pr domain.PullRequest
		var statusSmallint int
		if err := rows.Scan(
			&pr.PullRequestId,
			&pr.PullRequestName,
			&pr.AuthorId,
			&statusSmallint,
			&pr.CreatedAt,
			&pr.MergedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pr row: %w", err)
		}
		pr.Status = toDomainPullRequestStatus(statusSmallint)
		found[pr.PullRequestId] = pr
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// preserve order of input ids: append only those found
	out := make([]domain.PullRequest, 0, len(found))
	for _, id := range ids {
		if pr, ok := found[id]; ok {
			out = append(out, pr)
		}
	}

	return out, nil
}

func (r *PullRequestRepo) SetMerged(
	ctx context.Context,
	pullRequestId uuid.UUID,
) (domain.PullRequest, error) {
	sql, args, err := r.Builder.
		Update("pull_requests").
		Set("pr_status", 1).
		Set("merged_at", time.Now()).
		Where(squirrel.Eq{"id": pullRequestId}).
		Suffix("RETURNING id, pr_name, author_id, pr_status, created_at, merged_at").
		ToSql()
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("build update PR sql: %w", err)
	}

	var pr domain.PullRequest
	var statusSmallint int
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&pr.PullRequestId,
		&pr.PullRequestName,
		&pr.AuthorId,
		&statusSmallint,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.PullRequest{}, repoerrors.ErrNotFound
		}
		return domain.PullRequest{}, fmt.Errorf("exec update PR: %w", err)
	}

	pr.Status = toDomainPullRequestStatus(statusSmallint)

	return pr, nil
}
