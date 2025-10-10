//go:build integration

package integration

import (
	"context"
	"li-acc/internal/repository"
	"li-acc/internal/repository/db"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testRepo  *repository.Repository
	testDsn   string
	setupOnce sync.Once
	cleanup   func()
)

const MigrationsDir = "../db/migrations"

// --- Shared initializer (runs only once, even in partial test runs) ---
func ensureDBReady(t *testing.T) {
	t.Helper()

	setupOnce.Do(func() {
		var err error
		testDsn, cleanup, err = setupPostgres(context.Background())
		require.NoError(t, err)

		testRepo, err = repository.ConnectDB(testDsn)
		require.NoError(t, err, "failed to connect to test DB")

		migAbs, err := filepath.Abs(MigrationsDir)
		require.NoError(t, err)

		applied, err := db.ApplyMigrations(testRepo.DB, migAbs)
		require.NoError(t, err)
		if applied {
			t.Log("migrations applied successfully")
		}
	})
}

func TestMain(m *testing.M) {
	code := m.Run() // run all tests

	if cleanup != nil {
		cleanup()
	}
	os.Exit(code)
}

func TestConnectDB_ApplyMigrations(t *testing.T) {
	ensureDBReady(t)
}
