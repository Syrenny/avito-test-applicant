package helpers

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tc "github.com/testcontainers/testcontainers-go"
	postgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var isolationMu sync.Mutex

type DBInstance struct {
	Pool      *pgxpool.Pool
	Container *postgres.PostgresContainer
	Getter    *trmpgx.CtxGetter
}

// SetupTestDB starts a PostgreSQL testcontainer, applies migrations, and returns a connected pgx pool.
func SetupTestDB(ctx context.Context, migrationsPath string) (*DBInstance, error) {
	pgContainer, err := postgres.Run(ctx,
		"docker.io/postgres:18",
		postgres.WithDatabase("test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		tc.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("run postgres container: %w", err)
	}

	uri, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, uri)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("pgxpool new: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("pgx ping: %w", err)
	}

	if migrationsPath != "" {
		if err := runMigrations(uri, migrationsPath); err != nil {
			pool.Close()
			pgContainer.Terminate(ctx)
			return nil, err
		}
	}

	return &DBInstance{
		Pool:      pool,
		Container: pgContainer,
		Getter:    trmpgx.DefaultCtxGetter,
	}, nil
}

// Teardown stops container and closes the pool.
func (db *DBInstance) Teardown(ctx context.Context) error {
	if db == nil {
		return nil
	}

	if db.Pool != nil {
		db.Pool.Close()
	}

	if db.Container != nil {
		if err := db.Container.Terminate(ctx); err != nil {
			return fmt.Errorf("terminate container: %w", err)
		}
	}

	return nil
}

func runMigrations(databaseURL, migrationsPath string) error {
	source := fmt.Sprintf("file://%s", migrationsPath)
	m, err := migrate.New(source, databaseURL)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

// resetTestDB truncates all public tables to provide isolation between tests.
func resetTestDB(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return truncateAllTables(ctx, pool)
}

func truncateAllTables(ctx context.Context, pool *pgxpool.Pool) error {
	rows, err := pool.Query(ctx, `select tablename from pg_tables where schemaname = 'public' and tablename not like 'pg_%' and tablename not like 'sql_%'`)
	if err != nil {
		return fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan table name: %w", err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate tables: %w", err)
	}

	if len(tables) == 0 {
		return nil
	}

	sort.Strings(tables)

	quoted := make([]string, 0, len(tables))
	for _, table := range tables {
		quoted = append(quoted, `"`+strings.ReplaceAll(table, "\"", "\"\"")+`"`)
	}

	stmt := fmt.Sprintf("truncate table %s restart identity cascade", strings.Join(quoted, ", "))
	if _, err := pool.Exec(ctx, stmt); err != nil {
		return fmt.Errorf("truncate tables: %w", err)
	}

	return nil
}

// WithTestDatabase ensures each test gets a clean database state before and after execution.
func WithTestDatabase(t *testing.T, pool *pgxpool.Pool, fn func(ctx context.Context, pool *pgxpool.Pool)) {
	t.Helper()

	isolationMu.Lock()
	defer isolationMu.Unlock()

	if err := resetTestDB(pool); err != nil {
		t.Fatalf("reset database before test: %v", err)
	}

	ctx := context.Background()

	defer func() {
		if err := resetTestDB(pool); err != nil {
			t.Fatalf("reset database after test: %v", err)
		}
	}()

	fn(ctx, pool)
}

// InsertTestUser seeds a single user row and returns its identifier.
func InsertTestUser(ctx context.Context, pool *pgxpool.Pool) (uuid.UUID, error) {
	var id uuid.UUID
	if err := pool.QueryRow(ctx, `insert into users default values returning id`).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("insert test user: %w", err)
	}
	return id, nil
}

// ResetTestDB exposes the reset helper for callers that need manual control.
func ResetTestDB(ctx context.Context, pool *pgxpool.Pool) error {
	return truncateAllTables(ctx, pool)
}

// ResolveMigrationsDir returns the absolute path to the repository migrations folder.
func ResolveMigrationsDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("cannot determine caller path")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "migrations"), nil
}
