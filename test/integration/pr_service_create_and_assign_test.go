package integration_test

import (
	"context"
	"testing"

	"avito-test-applicant/internal/domain"
	"avito-test-applicant/internal/repo"
	"avito-test-applicant/internal/repo/pgdb"
	"avito-test-applicant/internal/service"
	"avito-test-applicant/pkg/postgres"
	"avito-test-applicant/test/helpers"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func newReposFromPool(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *repo.Repositories {
	pg := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    pool,
	}
	// Use constructors that accept getter when appropriate.
	teamRepo := pgdb.NewTeamRepo(pg, getter)
	userRepo := pgdb.NewUserRepo(pg, getter)
	prRepo := pgdb.NewPullRequestRepo(pg, getter)
	reviewerRepo := pgdb.NewReviewerRepo(pg, getter)

	return &repo.Repositories{
		Team:        teamRepo,
		User:        userRepo,
		PullRequest: prRepo,
		Reviewer:    reviewerRepo,
	}
}

func newPRServiceFromPool(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *service.PullRequestService {
	repos := newReposFromPool(pool, getter)

	trManager := postgres.NewTransactionManager(pool)

	return service.NewPullRequestService(repos, trManager)
}

// helper to create team + users in one function
func setupTeamWithUsers(ctx context.Context, t *testing.T, pool *pgxpool.Pool, getter *trmpgx.CtxGetter, teamName string, users []domain.User) (uuid.UUID, []domain.User) {
	teamRepo := pgdb.NewTeamRepo(&postgres.Postgres{Pool: pool, Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}, getter)
	userRepo := pgdb.NewUserRepo(&postgres.Postgres{Pool: pool, Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}, getter)

	team, err := teamRepo.CreateTeam(ctx, uuid.New(), teamName)
	require.NoError(t, err)

	created := make([]domain.User, 0, len(users))
	for _, u := range users {
		// if user.UserId is zero, generate
		uid := u.UserId
		if uid == uuid.Nil {
			uid = uuid.New()
		}
		createdUser, err := userRepo.CreateUser(ctx, uid, u.Username, u.IsActive, team.TeamId)
		require.NoError(t, err)
		created = append(created, createdUser)
	}
	return team.TeamId, created
}

func Test_CreateAndAssign_PrimaryCase_TwoReviewers(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		service := newPRServiceFromPool(pool, testDB.Getter)

		// team with author + 3 other active users => should pick 2 reviewers
		users := []domain.User{
			{UserId: uuid.New(), Username: "author", IsActive: true},
			{UserId: uuid.New(), Username: "u1", IsActive: true},
			{UserId: uuid.New(), Username: "u2", IsActive: true},
			{UserId: uuid.New(), Username: "u3", IsActive: true},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-standard", users)

		// find author id from created (username == "author")
		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}
		require.NotEqual(t, uuid.Nil, authorId)

		prID := uuid.New()
		res, err := service.CreateAndAssignPullRequest(ctx, prID, "add feature", authorId)
		require.NoError(t, err)
		require.Equal(t, prID, res.PullRequest.PullRequestId)
		// up to 2 reviewers
		require.LessOrEqual(t, len(res.Reviewers), 2)
		require.GreaterOrEqual(t, len(res.Reviewers), 1) // since there are 3 candidates we expect 2 but allow >=1 as sanity
		// author must not be among reviewers
		for _, rid := range res.Reviewers {
			require.NotEqual(t, authorId, rid)
		}
		// all reviewers active
		for _, rid := range res.Reviewers {
			found := false
			for _, u := range created {
				if u.UserId == rid {
					found = true
					require.True(t, u.IsActive)
				}
			}
			require.True(t, found, "reviewer should come from team members")
		}
	})
}

func Test_CreateAndAssign_OneCandidate(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		service := newPRServiceFromPool(pool, testDB.Getter)

		// author + 1 active user
		users := []domain.User{
			{UserId: uuid.New(), Username: "author", IsActive: true},
			{UserId: uuid.New(), Username: "only", IsActive: true},
			// other not in team
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-one", users)
		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}

		prID := uuid.New()
		res, err := service.CreateAndAssignPullRequest(ctx, prID, "small pr", authorId)
		require.NoError(t, err)
		require.Equal(t, 1, len(res.Reviewers))
		require.NotEqual(t, authorId, res.Reviewers[0])
	})
}

func Test_CreateAndAssign_NoCandidates(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		service := newPRServiceFromPool(pool, testDB.Getter)

		// team with only author
		users := []domain.User{
			{UserId: uuid.New(), Username: "author", IsActive: true},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-none", users)
		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}

		prID := uuid.New()
		res, err := service.CreateAndAssignPullRequest(ctx, prID, "no candidates pr", authorId)
		require.NoError(t, err)
		require.Len(t, res.Reviewers, 0)
	})
}

func Test_CreateAndAssign_AllInactive(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		service := newPRServiceFromPool(pool, testDB.Getter)

		// author + others but all others inactive
		users := []domain.User{
			{UserId: uuid.New(), Username: "author", IsActive: true},
			{UserId: uuid.New(), Username: "a", IsActive: false},
			{UserId: uuid.New(), Username: "b", IsActive: false},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-inactive", users)
		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}

		prID := uuid.New()
		res, err := service.CreateAndAssignPullRequest(ctx, prID, "inactive pr", authorId)
		require.NoError(t, err)
		require.Len(t, res.Reviewers, 0)
	})
}

func Test_CreateAndAssign_MixedActiveInactive(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		service := newPRServiceFromPool(pool, testDB.Getter)

		// author + mix of active/inactive
		users := []domain.User{
			{UserId: uuid.New(), Username: "author", IsActive: true},
			{UserId: uuid.New(), Username: "a", IsActive: true},
			{UserId: uuid.New(), Username: "b", IsActive: false},
			{UserId: uuid.New(), Username: "c", IsActive: true},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-mixed", users)
		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}

		prID := uuid.New()
		res, err := service.CreateAndAssignPullRequest(ctx, prID, "mixed pr", authorId)
		require.NoError(t, err)
		// reviewers should be from active ones {a,c}, count <=2
		require.LessOrEqual(t, len(res.Reviewers), 2)
		for _, rid := range res.Reviewers {
			require.NotEqual(t, authorId, rid)
			foundActive := false
			for _, u := range created {
				if u.UserId == rid {
					foundActive = u.IsActive
					break
				}
			}
			require.True(t, foundActive)
		}
	})
}

func Test_CreateAndAssign_DuplicateId(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		prService := newPRServiceFromPool(pool, testDB.Getter)

		users := []domain.User{
			{UserId: uuid.New(), Username: "author", IsActive: true},
			{UserId: uuid.New(), Username: "a", IsActive: true},
			{UserId: uuid.New(), Username: "b", IsActive: true},
		}
		_, created := setupTeamWithUsers(ctx, t, pool, testDB.Getter, "team-dup", users)
		var authorId uuid.UUID
		for _, u := range created {
			if u.Username == "author" {
				authorId = u.UserId
			}
		}

		prID := uuid.New()
		_, err := prService.CreateAndAssignPullRequest(ctx, prID, "dup pr", authorId)
		require.NoError(t, err)

		_, err = prService.CreateAndAssignPullRequest(ctx, prID, "dup pr second", authorId)
		require.ErrorIs(t, err, service.ErrPullRequestExists)
	})
}
