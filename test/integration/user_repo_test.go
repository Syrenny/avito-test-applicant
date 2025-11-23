package integration_test

import (
	"context"
	"testing"

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

func newUserRepoFromPool(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *pgdb.UserRepo {
	pg := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    pool,
	}
	return pgdb.NewUserRepo(pg, getter)
}

// --- Создание пользователя и получение по ID ---
func TestUserRepo_CreateAndGet(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		teamRepo := newTeamRepoFromPool(pool, testDB.Getter)
		userRepo := newUserRepoFromPool(pool, testDB.Getter)

		team, err := teamRepo.CreateTeam(ctx, uuid.New(), "team-create-get")
		require.NoError(t, err)

		userId := uuid.New()
		username := "alice"
		user, err := userRepo.CreateUser(ctx, userId, username, false, team.TeamId)
		require.NoError(t, err)
		require.Equal(t, username, user.Username)

		got, err := userRepo.GetUserById(ctx, userId)
		require.NoError(t, err)
		require.Equal(t, userId, got.UserId)
		require.Equal(t, username, got.Username)
	})
}

// --- Уникальность username в рамках команды ---
func TestUserRepo_UniqueUsername(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		teamRepo := newTeamRepoFromPool(pool, testDB.Getter)
		userRepo := newUserRepoFromPool(pool, testDB.Getter)

		team, err := teamRepo.CreateTeam(ctx, uuid.New(), "team-unique")
		require.NoError(t, err)

		username := "bob"
		_, err = userRepo.CreateUser(ctx, uuid.New(), username, true, team.TeamId)
		require.NoError(t, err)

		// Повторное создание с тем же username должно падать
		_, err = userRepo.CreateUser(ctx, uuid.New(), username, true, team.TeamId)
		require.ErrorIs(t, err, repoerrors.ErrUsernameTakenInTeam)
	})
}

// --- SetIsActive и проверка через GetUserById ---
func TestUserRepo_SetIsActive(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		teamRepo := newTeamRepoFromPool(pool, testDB.Getter)
		userRepo := newUserRepoFromPool(pool, testDB.Getter)

		team, err := teamRepo.CreateTeam(ctx, uuid.New(), "team-activate")
		require.NoError(t, err)

		userId := uuid.New()
		user, err := userRepo.CreateUser(ctx, userId, "charlie", false, team.TeamId)
		require.NoError(t, err)

		user, err = userRepo.SetIsActive(ctx, userId, true)
		require.NoError(t, err)
		require.True(t, user.IsActive)

		user, err = userRepo.SetIsActive(ctx, userId, false)
		require.NoError(t, err)
		require.False(t, user.IsActive)
	})
}

// --- UpdateUser ---
func TestUserRepo_UpdateUser(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		teamRepo := newTeamRepoFromPool(pool, testDB.Getter)
		userRepo := newUserRepoFromPool(pool, testDB.Getter)

		team, err := teamRepo.CreateTeam(ctx, uuid.New(), "team-update")
		require.NoError(t, err)

		userId := uuid.New()
		user, err := userRepo.CreateUser(ctx, userId, "dave", false, team.TeamId)
		require.NoError(t, err)

		user.Username = "dave-updated"
		user.IsActive = true
		updated, err := userRepo.UpdateUser(ctx, user)
		require.NoError(t, err)
		require.Equal(t, "dave-updated", updated.Username)
		require.True(t, updated.IsActive)
	})
}

// --- GetUsersByTeam ---
func TestUserRepo_GetUsersByTeam(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		teamRepo := newTeamRepoFromPool(pool, testDB.Getter)
		userRepo := newUserRepoFromPool(pool, testDB.Getter)

		team, err := teamRepo.CreateTeam(ctx, uuid.New(), "team-all-users")
		require.NoError(t, err)

		userRepo.CreateUser(ctx, uuid.New(), "eve", true, team.TeamId)
		userRepo.CreateUser(ctx, uuid.New(), "frank", false, team.TeamId)

		users, err := userRepo.GetUsersByTeam(ctx, team.TeamId)
		require.NoError(t, err)
		require.Len(t, users, 2)

		foundEve := false
		foundFrank := false
		for _, u := range users {
			if u.Username == "eve" {
				foundEve = true
			}
			if u.Username == "frank" {
				foundFrank = true
			}
		}
		require.True(t, foundEve && foundFrank)
	})
}

func TestUserRepo_Errors(t *testing.T) {

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		repo := newUserRepoFromPool(pool, testDB.Getter)

		nonExistentID := uuid.New()

		// --- GetUserById для несуществующего пользователя ---
		_, err := repo.GetUserById(ctx, nonExistentID)
		require.ErrorIs(t, err, repoerrors.ErrNotFound)

		// --- SetIsActive для несуществующего пользователя ---
		_, err = repo.SetIsActive(ctx, nonExistentID, true)
		require.ErrorIs(t, err, repoerrors.ErrNotFound)

		// --- UpdateUser для несуществующего пользователя ---
		user := domain.User{
			UserId:   nonExistentID,
			Username: "ghost",
			IsActive: true,
		}
		_, err = repo.UpdateUser(ctx, user)
		require.ErrorIs(t, err, repoerrors.ErrNotFound)
	})
}
