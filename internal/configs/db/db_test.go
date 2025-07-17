package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB(t *testing.T) {
	tests := []struct {
		name      string
		driver    string
		dsn       string
		shouldErr bool
	}{
		{
			name:      "Valid SQLite in-memory",
			driver:    "sqlite",
			dsn:       "file::memory:?cache=shared",
			shouldErr: false,
		},
		{
			name:      "Invalid driver",
			driver:    "invalid-driver",
			dsn:       "file::memory:?cache=shared",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConn, err := NewDB(tt.driver, tt.dsn)
			if tt.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, dbConn)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, dbConn)
				_ = dbConn.Close()
			}
		})
	}
}

func TestRunMigrations(t *testing.T) {
	// Создаем временную директорию для миграций
	tmpDir := t.TempDir()

	// Создаем корректный SQL-файл с директивами goose
	migrationFile := filepath.Join(tmpDir, "00001_create_table.sql")
	err := os.WriteFile(migrationFile, []byte(`
-- +goose Up
CREATE TABLE test_table (
	id INTEGER PRIMARY KEY,
	name TEXT
);

-- +goose Down
DROP TABLE test_table;
`), 0644)
	require.NoError(t, err)

	dbConn, err := NewDB("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)
	defer dbConn.Close()

	t.Run("Run valid migration", func(t *testing.T) {
		err := RunMigrations(dbConn, "sqlite", tmpDir)
		assert.NoError(t, err)

		var exists int
		err = dbConn.Get(&exists, `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='test_table'`)
		assert.NoError(t, err)
		assert.Equal(t, 1, exists)
	})

	t.Run("Invalid path", func(t *testing.T) {
		err := RunMigrations(dbConn, "sqlite", "/non/existing/path")
		assert.Error(t, err)
	})

	t.Run("Unsupported dialect", func(t *testing.T) {
		err := RunMigrations(dbConn, "invalid-dialect", tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown dialect")
	})
}
