package integration_test
import (
	"context"
	"testing"

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

func newReviewerRepoFromPool(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *pgdb.ReviewerRepo {
	pg := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    pool,
	}
	return pgdb.NewReviewerRepo(pg, getter)
}

func TestReviewerRepo_AssignAndList(t *testing.T) {
	t.Parallel()

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		userRepo := newUserRepoFromPool(pool, testDB.Getter)
		prRepo := newPullRequestRepoFromPool(pool, testDB.Getter) // твой PR репо
		reviewerRepo := newReviewerRepoFromPool(pool, testDB.Getter)

		// --- Подготовка пользователей и PR ---
		teamRepo := newTeamRepoFromPool(pool, testDB.Getter)
		team, err := teamRepo.CreateTeam(ctx, uuid.New(), "team-review")
		require.NoError(t, err)

		user1, err := userRepo.CreateUser(ctx, uuid.New(), "alice", true, team.TeamId)
		require.NoError(t, err)
		_, err = userRepo.CreateUser(ctx, uuid.New(), "bob", true, team.TeamId)
		require.NoError(t, err)

		pr, err := prRepo.CreatePullRequest(ctx, uuid.New(), "PR 1", user1.UserId)
		require.NoError(t, err)

		// --- Назначение одного ревьюера и проверка ListReviewers ---
		err = reviewerRepo.AssignOne(ctx, pr.PullRequestId, user1.UserId)
		require.NoError(t, err)

		reviewers, err := reviewerRepo.ListReviewers(ctx, pr.PullRequestId)
		require.NoError(t, err)
		require.Len(t, reviewers, 1)
		require.Contains(t, reviewers, user1.UserId)
	})
}

func TestReviewerRepo_AssignAndRemove(t *testing.T) {
	t.Parallel()

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		userRepo := newUserRepoFromPool(pool, testDB.Getter)
		prRepo := newPullRequestRepoFromPool(pool, testDB.Getter)
		reviewerRepo := newReviewerRepoFromPool(pool, testDB.Getter)

		teamRepo := newTeamRepoFromPool(pool, testDB.Getter)
		team, _ := teamRepo.CreateTeam(ctx, uuid.New(), "team-remove")

		user, _ := userRepo.CreateUser(ctx, uuid.New(), "charlie", true, team.TeamId)
		pr, _ := prRepo.CreatePullRequest(ctx, uuid.New(), "PR 2", user.UserId)

		// Назначаем и проверяем
		err := reviewerRepo.AssignOne(ctx, pr.PullRequestId, user.UserId)
		require.NoError(t, err)

		reviewers, _ := reviewerRepo.ListReviewers(ctx, pr.PullRequestId)
		require.Contains(t, reviewers, user.UserId)

		// Удаляем и проверяем
		err = reviewerRepo.RemoveOne(ctx, pr.PullRequestId, user.UserId)
		require.NoError(t, err)

		reviewers, _ = reviewerRepo.ListReviewers(ctx, pr.PullRequestId)
		require.NotContains(t, reviewers, user.UserId)
	})
}

func TestReviewerRepo_MultipleReviewers(t *testing.T) {
	t.Parallel()

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		userRepo := newUserRepoFromPool(pool, testDB.Getter)
		prRepo := newPullRequestRepoFromPool(pool, testDB.Getter)
		reviewerRepo := newReviewerRepoFromPool(pool, testDB.Getter)

		team, _ := newTeamRepoFromPool(pool, testDB.Getter).CreateTeam(ctx, uuid.New(), "team-multi")
		user1, _ := userRepo.CreateUser(ctx, uuid.New(), "alice", true, team.TeamId)
		user2, _ := userRepo.CreateUser(ctx, uuid.New(), "bob", true, team.TeamId)
		user3, _ := userRepo.CreateUser(ctx, uuid.New(), "charlie", true, team.TeamId)

		pr, _ := prRepo.CreatePullRequest(ctx, uuid.New(), "PR multi", user1.UserId)

		// Назначаем нескольких
		for _, u := range []uuid.UUID{user1.UserId, user2.UserId, user3.UserId} {
			err := reviewerRepo.AssignOne(ctx, pr.PullRequestId, u)
			require.NoError(t, err)
		}

		reviewers, err := reviewerRepo.ListReviewers(ctx, pr.PullRequestId)
		require.NoError(t, err)
		require.Len(t, reviewers, 3)
		require.ElementsMatch(t, []uuid.UUID{user1.UserId, user2.UserId, user3.UserId}, reviewers)
	})
}

func TestReviewerRepo_ListByUserId(t *testing.T) {
	t.Parallel()

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		userRepo := newUserRepoFromPool(pool, testDB.Getter)
		prRepo := newPullRequestRepoFromPool(pool, testDB.Getter)
		reviewerRepo := newReviewerRepoFromPool(pool, testDB.Getter)

		team, _ := newTeamRepoFromPool(pool, testDB.Getter).CreateTeam(ctx, uuid.New(), "team-list")
		user1, _ := userRepo.CreateUser(ctx, uuid.New(), "alice", true, team.TeamId)
		user2, _ := userRepo.CreateUser(ctx, uuid.New(), "bob", true, team.TeamId)

		pr1, _ := prRepo.CreatePullRequest(ctx, uuid.New(), "PR A", user1.UserId)
		pr2, _ := prRepo.CreatePullRequest(ctx, uuid.New(), "PR B", user1.UserId)
		pr3, _ := prRepo.CreatePullRequest(ctx, uuid.New(), "PR C", user2.UserId)

		// Назначаем
		reviewerRepo.AssignOne(ctx, pr1.PullRequestId, user1.UserId)
		reviewerRepo.AssignOne(ctx, pr2.PullRequestId, user1.UserId)
		reviewerRepo.AssignOne(ctx, pr3.PullRequestId, user2.UserId)

		prsUser1, err := reviewerRepo.ListByUserId(ctx, user1.UserId)
		require.NoError(t, err)
		require.Len(t, prsUser1, 2)
		require.ElementsMatch(t, []uuid.UUID{pr1.PullRequestId, pr2.PullRequestId}, prsUser1)

		prsUser2, err := reviewerRepo.ListByUserId(ctx, user2.UserId)
		require.NoError(t, err)
		require.Len(t, prsUser2, 1)
		require.Equal(t, pr3.PullRequestId, prsUser2[0])
	})
}

func TestReviewerRepo_Errors(t *testing.T) {
	t.Parallel()

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		reviewerRepo := newReviewerRepoFromPool(pool, testDB.Getter)

		nonExistentPR := uuid.New()
		nonExistentUser := uuid.New()

		// Удаление несуществующей связи
		err := reviewerRepo.RemoveOne(ctx, nonExistentPR, nonExistentUser)
		require.ErrorIs(t, err, repoerrors.ErrNotFound)

	})
}
