package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	sqliteDriver = "sqlite"
	sqliteDSN    = ":memory:"
)

// createTempMigrationDir creates a temporary directory and a valid Goose migration SQL file.
func createTempMigrationDir(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	migration := `
-- +goose Up
CREATE TABLE test_table (id INTEGER PRIMARY KEY);

-- +goose Down
DROP TABLE test_table;
`

	file := filepath.Join(tmpDir, "00001_create_test_table.sql")
	if err := os.WriteFile(file, []byte(migration), 0644); err != nil {
		t.Fatalf("failed to write migration file: %v", err)
	}

	return tmpDir
}

func TestNewDB_Success(t *testing.T) {
	db, err := NewDB("sqlite", ":memory:")
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer db.Close()
}

func TestNewDB_Failure(t *testing.T) {
	// Using invalid driver should produce error
	db, err := NewDB("invalid-driver", "some-dsn")
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestNewDB(t *testing.T) {
	db, err := NewDB(
		sqliteDriver,
		sqliteDSN,
		WithMaxOpenConns(5),
		WithMaxIdleConns(2),
		WithConnMaxLifetime(10*time.Minute),
	)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Validate we can ping
	err = db.Ping()
	assert.NoError(t, err)

	// You could also assert settings here if needed:
	// but note: sqlx does not expose GetMaxOpenConns, etc.
	_ = db.Close()
}

func TestRunMigrations(t *testing.T) {
	db, err := NewDB(sqliteDriver, sqliteDSN)
	assert.NoError(t, err)
	defer db.Close()

	migrationsDir := createTempMigrationDir(t)

	err = RunMigrations(db, sqliteDriver, migrationsDir)
	assert.NoError(t, err)

	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_table';").Scan(&tableName)
	assert.NoError(t, err)
	assert.Equal(t, "test_table", tableName)
}

func TestRunMigrations_InvalidDialect(t *testing.T) {
	db, err := NewDB(sqliteDriver, sqliteDSN)
	assert.NoError(t, err)
	defer db.Close()

	err = RunMigrations(db, "invalid-dialect", "some/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown dialect")
}

func TestRunMigrations_InvalidPath(t *testing.T) {
	db, err := NewDB(sqliteDriver, sqliteDSN)
	assert.NoError(t, err)
	defer db.Close()

	err = RunMigrations(db, sqliteDriver, "/non/existent/migrations")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory does not exist")
}

func TestRunMigrations_BrokenSQL(t *testing.T) {
	db, err := NewDB(sqliteDriver, sqliteDSN)
	assert.NoError(t, err)
	defer db.Close()

	tmpDir := t.TempDir()
	brokenMigration := `
-- +goose Up
THIS IS NOT VALID SQL;

-- +goose Down
DROP TABLE dummy;
`
	file := filepath.Join(tmpDir, "00001_broken.sql")
	err = os.WriteFile(file, []byte(brokenMigration), 0644)
	assert.NoError(t, err)

	err = RunMigrations(db, sqliteDriver, tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "syntax error") // or general "near" depending on SQLite message
}
