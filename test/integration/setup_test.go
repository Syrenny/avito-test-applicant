package integration_test

import (
	"context"
	"os"
	"testing"

	"avito-test-applicant/test/helpers"
)

var (
	testDB *helpers.DBInstance
)

// TestMain provides process-wide setup/teardown for integration tests.
func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	migrationsDir, err := helpers.ResolveMigrationsDir()
	if err != nil {
		panic(err)
	}

	testDB, err = helpers.SetupTestDB(ctx, migrationsDir)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	if testDB != nil {
		if err := testDB.Teardown(context.Background()); err != nil {
			panic(err)
		}
	}

	os.Exit(code)
}
