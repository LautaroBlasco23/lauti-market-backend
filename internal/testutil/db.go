//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDB spins up a PostgreSQL container, runs all migrations, and returns
// a connected *sql.DB. The container is terminated on test cleanup.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	require.NoError(t, db.PingContext(ctx))

	runMigrations(t, db)

	return db
}

// TruncateTables truncates all application tables between tests (preserving schema).
func TruncateTables(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE TABLE payments, order_items, orders, products, auths, stores, users RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func runMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	// Find migrations directory relative to this file's module root.
	migrationsDir := findMigrationsDir(t)

	entries, err := os.ReadDir(migrationsDir)
	require.NoError(t, err)

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, filepath.Join(migrationsDir, e.Name()))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		data, err := os.ReadFile(f)
		require.NoError(t, err, "reading migration %s", f)
		_, err = db.Exec(string(data))
		require.NoError(t, err, "running migration %s", f)
	}
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()
	// Walk up from the current working directory to find the migrations folder.
	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		candidate := filepath.Join(dir, "migrations")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find migrations directory")
		}
		dir = parent
	}
}
