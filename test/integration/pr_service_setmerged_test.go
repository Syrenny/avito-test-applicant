package integration_test

import (
	"context"
	"testing"
	"time"

	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/service"
	"avito-test-applicant/test/helpers"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func Test_SetMerged_MarksPRMergedAndSetsMergedAt(t *testing.T) {
	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		// arrange
		prService := newPRServiceFromPool(pool, testDB.Getter)

		users := []domain.User{
			{Username: "author", IsActive: true},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-merge-1", users)

		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}
		require.NotEqual(t, uuid.Nil, authorId)

		prID := uuid.New()
		_, err := prService.CreateAndAssignPullRequest(ctx, prID, "merge-test", authorId)
		require.NoError(t, err)

		// act
		merged, err := prService.SetMerged(ctx, prID)

		// assert
		require.NoError(t, err)
		require.Equal(t, domain.PullRequestStatusMERGED, merged.Status)
		require.NotNil(t, merged.MergedAt)
		// mergedAt should be roughly now
		require.WithinDuration(t, time.Now(), *merged.MergedAt, time.Second*5)
	})
}

func Test_SetMerged_PreventsReviewerChangesAfterMerge(t *testing.T) {
	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		prService := newPRServiceFromPool(pool, testDB.Getter)
		reviewerRepo := newReviewerRepoFromPool(pool, testDB.Getter)

		// arrange: author + two reviewers + one candidate
		users := []domain.User{
			{Username: "author", IsActive: true},
			{Username: "r1", IsActive: true},
			{Username: "r2", IsActive: true},
			{Username: "candidate", IsActive: true},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-merge-2", users)

		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}
		require.NotEqual(t, uuid.Nil, authorId)

		prID := uuid.New()
		_, err := prService.CreateAndAssignPullRequest(ctx, prID, "merge-prevent-changes", authorId)
		require.NoError(t, err)

		// read current reviewers (before merge)
		before, err := reviewerRepo.ListReviewers(ctx, prID)
		require.NoError(t, err)

		// merge
		merged, err := prService.SetMerged(ctx, prID)
		require.NoError(t, err)
		require.Equal(t, domain.PullRequestStatusMERGED, merged.Status)

		// attempt reassign — should error with ErrPullRequestMerged
		if len(before) == 0 {
			// если по каким-то причинам ревьюверов нет — попытка реассайна какого-либо пользователя, не автор
			var some uuid.UUID
			for _, u := range created {
				if u.UserId != authorId {
					some = u.UserId
					break
				}
			}
			_, err = prService.Reassign(ctx, prID, some)
		} else {
			_, err = prService.Reassign(ctx, prID, before[0])
		}
		require.ErrorIs(t, err, service.ErrPullRequestMerged)

		// reviewers should remain unchanged
		after, err := reviewerRepo.ListReviewers(ctx, prID)
		require.NoError(t, err)
		require.ElementsMatch(t, before, after)
	})
}

func Test_SetMerged_Idempotent(t *testing.T) {
	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		prService := newPRServiceFromPool(pool, testDB.Getter)

		users := []domain.User{
			{Username: "author", IsActive: true},
			{Username: "r1", IsActive: true},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-merge-3", users)

		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}
		require.NotEqual(t, uuid.Nil, authorId)

		prID := uuid.New()
		_, err := prService.CreateAndAssignPullRequest(ctx, prID, "merge-idempotent", authorId)
		require.NoError(t, err)

		// first merge
		m1, err := prService.SetMerged(ctx, prID)
		require.NoError(t, err)
		require.Equal(t, domain.PullRequestStatusMERGED, m1.Status)
		require.NotNil(t, m1.MergedAt)

		// second merge - should be idempotent (no error, same state)
		m2, err := prService.SetMerged(ctx, prID)
		require.NoError(t, err)
		require.Equal(t, domain.PullRequestStatusMERGED, m2.Status)
		require.NotNil(t, m2.MergedAt)

		// mergedAt should be set (non-nil) — it may be equal or very close in time
		require.WithinDuration(t, *m1.MergedAt, *m2.MergedAt, time.Second)
	})
}
