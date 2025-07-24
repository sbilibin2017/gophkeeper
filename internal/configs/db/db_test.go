package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func TestNewDB(t *testing.T) {
	dsn := ":memory:"
	driver := "sqlite"

	conn, err := New(driver, dsn)
	require.NoError(t, err)
	require.NotNil(t, conn)

	err = conn.Ping()
	assert.NoError(t, err)
}

func TestWithMaxOpenConns(t *testing.T) {
	dsn := ":memory:"
	driver := "sqlite"

	conn, err := New(driver, dsn, WithMaxOpenConns(7))
	require.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestWithMaxIdleConns(t *testing.T) {
	dsn := ":memory:"
	driver := "sqlite"

	conn, err := New(driver, dsn, WithMaxIdleConns(4))
	require.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestWithConnMaxLifetime(t *testing.T) {
	dsn := ":memory:"
	driver := "sqlite"

	conn, err := New(driver, dsn, WithConnMaxLifetime(30*time.Second))
	require.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestMultipleOptions(t *testing.T) {
	dsn := ":memory:"
	driver := "sqlite"

	conn, err := New(driver, dsn,
		WithMaxOpenConns(20),
		WithMaxIdleConns(5),
		WithConnMaxLifetime(1*time.Minute),
	)
	require.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestRunMigrations(t *testing.T) {
	// Step 1: Create a temp file DB
	tmpDB, err := os.CreateTemp("", "testdb-*.sqlite")
	require.NoError(t, err)
	defer os.Remove(tmpDB.Name())

	conn, err := New("sqlite", tmpDB.Name())
	require.NoError(t, err)
	defer conn.Close()

	// Step 2: Create temp migrations dir with 1 migration
	migrationsDir := t.TempDir()

	migrationContent := `
-- +goose Up
CREATE TABLE test_table (
    id INTEGER PRIMARY KEY,
    name TEXT
);

-- +goose Down
DROP TABLE test_table;
`
	migrationPath := filepath.Join(migrationsDir, "00001_create_test_table.sql")
	err = os.WriteFile(migrationPath, []byte(migrationContent), 0644)
	require.NoError(t, err)

	// Step 3: Run migrations
	err = RunMigrations(conn, "sqlite", migrationsDir)
	require.NoError(t, err)

	// Step 4: Verify table exists
	var tableName string
	err = conn.Get(&tableName, `SELECT name FROM sqlite_master WHERE type='table' AND name='test_table'`)
	assert.NoError(t, err)
	assert.Equal(t, "test_table", tableName)
}
