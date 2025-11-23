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

func newTeamRepoFromPool(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *pgdb.TeamRepo {
	pg := &postgres.Postgres{
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		Pool:    pool,
	}
	return pgdb.NewTeamRepo(pg, getter)
}

func TestTeamRepo_CreateAndGet(t *testing.T) {
	t.Parallel()

	helpers.WithTestDatabase(t, testDB.Pool, func(ctx context.Context, pool *pgxpool.Pool) {
		repo := newTeamRepoFromPool(pool, testDB.Getter)

		teamName := "test-team"
		teamId, _ := uuid.Parse("11111111-1111-1111-1111-111111111111")

		// Создание команды
		team, err := repo.CreateTeam(ctx, teamId, teamName)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, team.TeamId)
		require.Equal(t, teamName, team.TeamName)

		// Повторная попытка создания должна выдавать ErrAlreadyExists
		_, err = repo.CreateTeam(ctx, teamId, teamName)
		require.ErrorIs(t, err, repoerrors.ErrAlreadyExists)

		// Получение по имени
		got, err := repo.GetTeamByName(ctx, teamName)
		require.NoError(t, err)
		require.Equal(t, team.TeamId, got.TeamId)
		require.Equal(t, team.TeamName, got.TeamName)

		// Получение несуществующей команды по имени
		_, err = repo.GetTeamByName(ctx, "non-existent")
		require.ErrorIs(t, err, repoerrors.ErrNotFound)

		// Получение по ID
		got, err = repo.GetTeamById(ctx, team.TeamId)
		require.NoError(t, err)
		require.Equal(t, team.TeamName, got.TeamName)

		// Получение несуществующей команды по ID
		_, err = repo.GetTeamById(ctx, uuid.New())
		require.ErrorIs(t, err, repoerrors.ErrNotFound)
	})
}
