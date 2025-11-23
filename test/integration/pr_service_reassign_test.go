package integration_test

import (
	"context"
	"testing"
	"time"

	"avito-test-applicant/internal/domain"
	"avito-test-applicant/test/helpers"

	"avito-test-applicant/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func Test_Reassign_StandardCase(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		service := newPRServiceFromPool(pool, testDB.Getter)

		users := []domain.User{
			{Username: "author", IsActive: true},
			{Username: "rev1", IsActive: true},
			{Username: "rev2", IsActive: true},
			{Username: "candidate1", IsActive: true},
			{Username: "candidate2", IsActive: true},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-standard", users)

		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}

		prID := uuid.New()
		prRes, err := service.CreateAndAssignPullRequest(ctx, prID, "feature pr", authorId)
		require.NoError(t, err)
		require.Len(t, prRes.Reviewers, 2)

		// Переназначаем первого ревьювера
		oldUser := prRes.Reviewers[0]
		res, err := service.Reassign(ctx, prID, oldUser)
		require.NoError(t, err)
		require.Equal(t, prID, res.PullRequest.PullRequestId)
		require.Len(t, res.Reviewers, 2)

		// автор не среди ревьюверов
		for _, uid := range res.Reviewers {
			require.NotEqual(t, authorId, uid)
		}

		// новый ревьювер активен и отличается от oldUser
		found := false
		for _, u := range created {
			for _, r := range res.Reviewers {
				if r == u.UserId && r != oldUser {
					require.True(t, u.IsActive)
					found = true
				}
			}
		}
		require.True(t, found, "replacement reviewer must be active and different from oldUser")
	})
}

func Test_Reassign_NoCandidates(t *testing.T) {
	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		prService := newPRServiceFromPool(pool, testDB.Getter)

		users := []domain.User{
			{Username: "author", IsActive: true},
			{Username: "rev1", IsActive: true},
			{Username: "rev2", IsActive: true},
			{Username: "candidate", IsActive: false},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-nocand", users)
		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}

		prID := uuid.New()
		prRes, err := prService.CreateAndAssignPullRequest(ctx, prID, "pr no candidates", authorId)
		require.NoError(t, err)
		require.Len(t, prRes.Reviewers, 2)

		oldUser := prRes.Reviewers[0]
		_, err = prService.Reassign(ctx, prID, oldUser)
		require.ErrorIs(t, err, service.ErrNoCandidate)
	})
}

func Test_Reassign_OnlyOldUserInTeam(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		prService := newPRServiceFromPool(pool, testDB.Getter)

		users := []domain.User{
			{Username: "author", IsActive: true},
			{Username: "rev1", IsActive: true}, // единственный в команде
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-one-member", users)
		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}

		prID := uuid.New()
		prRes, err := prService.CreateAndAssignPullRequest(ctx, prID, "single user team pr", authorId)
		require.NoError(t, err)
		require.Len(t, prRes.Reviewers, 1)

		oldUser := prRes.Reviewers[0]
		_, err = prService.Reassign(ctx, prID, oldUser)
		require.ErrorIs(t, err, service.ErrNoCandidate)
	})
}

func Test_Reassign_Randomness(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		service := newPRServiceFromPool(pool, testDB.Getter)

		users := []domain.User{
			{Username: "author", IsActive: true},
			{Username: "rev1", IsActive: true},
			{Username: "rev2", IsActive: true},
			{Username: "c1", IsActive: true},
			{Username: "c2", IsActive: true},
			{Username: "c3", IsActive: true}, // добавлено для стабильной замены
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-random", users)

		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}

		prID := uuid.New()
		prRes, err := service.CreateAndAssignPullRequest(ctx, prID, "pr random", authorId)
		require.NoError(t, err)
		require.Len(t, prRes.Reviewers, 2)

		oldUser := prRes.Reviewers[0]

		var replacedIds []uuid.UUID
		// несколько повторных попыток замены, чтобы проверить рандомизацию
		for i := 0; i < 5; i++ {
			currentOldUser := oldUser
			for _, r := range prRes.Reviewers {
				if r != currentOldUser {
					currentOldUser = r
					break
				}
			}

			res, err := service.Reassign(ctx, prID, currentOldUser)
			require.NoError(t, err)
			require.Len(t, res.Reviewers, 2)
			replacedIds = append(replacedIds, res.Reviewers...)
			prRes.Reviewers = res.Reviewers
		}

		var foundDifferent bool
		for _, rid := range replacedIds {
			if rid != oldUser && rid != authorId {
				foundDifferent = true
				break
			}
		}
		require.True(t, foundDifferent, "at least one reassigned reviewer must differ from oldUser and author")
	})
}

func Test_Reassign_AfterMerged_Forbidden(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		svc := newPRServiceFromPool(pool, testDB.Getter)

		// create team: author + two reviewers + one candidate
		users := []domain.User{
			{Username: "author", IsActive: true},
			{Username: "r1", IsActive: true},
			{Username: "r2", IsActive: true},
			{Username: "c1", IsActive: true},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-merged", users)

		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}
		require.NotEqual(t, uuid.Nil, authorId)

		prID := uuid.New()
		prRes, err := svc.CreateAndAssignPullRequest(ctx, prID, "pr-to-merge", authorId)
		require.NoError(t, err)

		// merge it
		merged, err := svc.SetMerged(ctx, prID)
		require.NoError(t, err)
		require.Equal(t, domain.PullRequestStatusMERGED, merged.Status)
		require.NotNil(t, merged.MergedAt)
		require.WithinDuration(t, time.Now(), *merged.MergedAt, time.Second*2)

		// attempt reassign -> forbidden
		if len(prRes.Reviewers) == 0 {
			// if no reviewers were assigned for some reason, pick some user that isn't author
			var fallback uuid.UUID
			for _, u := range created {
				if u.UserId != authorId {
					fallback = u.UserId
					break
				}
			}
			_, err = svc.Reassign(ctx, prID, fallback)
		} else {
			_, err = svc.Reassign(ctx, prID, prRes.Reviewers[0])
		}
		require.ErrorIs(t, err, service.ErrPullRequestMerged)
	})
}

func Test_Reassign_NonExistentPR(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		svc := newPRServiceFromPool(pool, testDB.Getter)

		_, err := svc.Reassign(ctx, uuid.New(), uuid.New())
		require.ErrorIs(t, err, service.ErrPullRequestNotFound)
	})
}

func Test_Reassign_OldUserNotAssigned(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		svc := newPRServiceFromPool(pool, testDB.Getter)

		// team: author + r1 + r2 + candidate
		users := []domain.User{
			{Username: "author", IsActive: true},
			{Username: "r1", IsActive: true},
			{Username: "r2", IsActive: true},
			{Username: "candidate", IsActive: true},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-not-assigned", users)
		var authorId uuid.UUID
		var notAssigned uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
			if u.Username == "candidate" {
				notAssigned = u.UserId
			}
		}

		prID := uuid.New()
		prRes, err := svc.CreateAndAssignPullRequest(ctx, prID, "pr-no-old", authorId)
		require.NoError(t, err)

		// choose a user that is NOT assigned (candidate is not guaranteed to be unassigned, but we check)
		// ensure chosen user actually not in assigned list
		chosen := notAssigned
		assignedMap := map[uuid.UUID]struct{}{}
		for _, id := range prRes.Reviewers {
			assignedMap[id] = struct{}{}
		}
		// if candidate is assigned accidentally, pick another user not assigned
		if _, ok := assignedMap[chosen]; ok {
			for _, u := range created {
				if u.UserId == authorId {
					continue
				}
				if _, inAssigned := assignedMap[u.UserId]; !inAssigned {
					chosen = u.UserId
					break
				}
			}
		}

		// now chosen should not be assigned
		_, err = svc.Reassign(ctx, prID, chosen)
		require.ErrorIs(t, err, service.ErrUserNotFound)
	})
}
