package integration_test

import (
	"context"
	"testing"
	"time"

	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo/pgdb"
	"avito-test-applicant/internal/repo/repoerrors"
	"avito-test-applicant/pkg/postgres"
	"avito-test-applicant/test/helpers"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func newPullRequestRepoFromPool(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *pgdb.PullRequestRepo {
	pg := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    pool,
	}
	return pgdb.NewPullRequestRepo(pg, getter)
}

func TestPullRequestRepo_CreateAndGet(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		userRepo := newUserRepoFromPool(pool, testDB.Getter)
		prRepo := newPullRequestRepoFromPool(pool, testDB.Getter)

		team, _ := newTeamRepoFromPool(pool, testDB.Getter).CreateTeam(ctx, uuid.New(), "team-pr-create")
		author, _ := userRepo.CreateUser(ctx, uuid.New(), "alice", true, team.TeamId)

		prId := uuid.New()
		prName := "PR 1"
		pr, err := prRepo.CreatePullRequest(ctx, prId, prName, author.UserId)
		require.NoError(t, err)
		require.Equal(t, prId, pr.PullRequestId)
		require.Equal(t, prName, pr.PullRequestName)
		require.Equal(t, author.UserId, pr.AuthorId)
		require.Equal(t, domain.PullRequestStatusOPEN, pr.Status)

		got, err := prRepo.GetPullRequestById(ctx, prId)
		require.NoError(t, err)
		require.Equal(t, pr.PullRequestId, got.PullRequestId)
	})
}

func TestPullRequestRepo_CreateAndGetByIds(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		userRepo := newUserRepoFromPool(pool, testDB.Getter)
		prRepo := newPullRequestRepoFromPool(pool, testDB.Getter)

		team, _ := newTeamRepoFromPool(pool, testDB.Getter).CreateTeam(ctx, uuid.New(), "team-pr-multi")
		author, _ := userRepo.CreateUser(ctx, uuid.New(), "bob", true, team.TeamId)

		prs := make([]domain.PullRequest, 0, 3)
		ids := make([]uuid.UUID, 0, 3)
		for i := 1; i <= 3; i++ {
			pr, _ := prRepo.CreatePullRequest(ctx, uuid.New(), "PR "+string(rune(i+'A'-1)), author.UserId)
			prs = append(prs, pr)
			ids = append(ids, pr.PullRequestId)
		}

		gotPRs, err := prRepo.GetPullRequestsByIds(ctx, ids)
		require.NoError(t, err)
		require.Len(t, gotPRs, 3)
		for _, pr := range prs {
			found := false
			for _, got := range gotPRs {
				if got.PullRequestId == pr.PullRequestId {
					found = true
					break
				}
			}
			require.True(t, found)
		}
	})
}

func TestPullRequestRepo_SetMerged(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		userRepo := newUserRepoFromPool(pool, testDB.Getter)
		prRepo := newPullRequestRepoFromPool(pool, testDB.Getter)

		team, _ := newTeamRepoFromPool(pool, testDB.Getter).CreateTeam(ctx, uuid.New(), "team-pr-merged")
		author, _ := userRepo.CreateUser(ctx, uuid.New(), "charlie", true, team.TeamId)

		pr, _ := prRepo.CreatePullRequest(ctx, uuid.New(), "PR Merge", author.UserId)
		require.Equal(t, domain.PullRequestStatusOPEN, pr.Status)

		merged, err := prRepo.SetMerged(ctx, pr.PullRequestId)
		require.NoError(t, err)
		require.Equal(t, domain.PullRequestStatusMERGED, merged.Status)
		require.WithinDuration(t, time.Now(), *merged.MergedAt, time.Second)
	})
}

func TestPullRequestRepo_Errors(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		prRepo := newPullRequestRepoFromPool(pool, testDB.Getter)

		nonExistentID := uuid.New()

		// GetPullRequestById для несуществующего PR
		_, err := prRepo.GetPullRequestById(ctx, nonExistentID)
		require.ErrorIs(t, err, repoerrors.ErrNotFound)

		// SetMerged для несуществующего PR
		_, err = prRepo.SetMerged(ctx, nonExistentID)
		require.ErrorIs(t, err, repoerrors.ErrNotFound)

		// GetPullRequestsByIds с пустым списком
		prs, err := prRepo.GetPullRequestsByIds(ctx, []uuid.UUID{})
		require.NoError(t, err)
		require.Len(t, prs, 0)
	})
}

func TestPullRequestRepo_CreateAlreadyExists(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		userRepo := newUserRepoFromPool(pool, testDB.Getter)
		prRepo := newPullRequestRepoFromPool(pool, testDB.Getter)

		team, _ := newTeamRepoFromPool(pool, testDB.Getter).CreateTeam(ctx, uuid.New(), "team-pr-duplicate")
		author, _ := userRepo.CreateUser(ctx, uuid.New(), "dave", true, team.TeamId)

		prId := uuid.New()
		_, err := prRepo.CreatePullRequest(ctx, prId, "PR Original", author.UserId)
		require.NoError(t, err)

		// Повторная попытка с тем же ID
		_, err = prRepo.CreatePullRequest(ctx, prId, "PR Duplicate", author.UserId)
		require.ErrorIs(t, err, repoerrors.ErrAlreadyExists)
	})
}
