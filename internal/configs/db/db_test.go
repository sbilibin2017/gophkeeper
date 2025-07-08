package db

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDB(t *testing.T) {
	// Get absolute path to the db.sqlite file created by NewDB
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "failed to get runtime caller info")

	dir := filepath.Dir(filename)
	dbPath := filepath.Join(dir, "db.sqlite")

	// Cleanup before test: remove existing DB file if any
	_ = os.Remove(dbPath)

	// Call the function under test
	db, err := NewDB()
	require.NoError(t, err)
	require.NotNil(t, db)

	// Ping DB to ensure it's working
	err = db.Ping()
	require.NoError(t, err)

	// Run a simple query to confirm DB works
	var result string
	err = db.Get(&result, "PRAGMA journal_mode;")
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Close DB connection before cleanup
	err = db.Close()
	require.NoError(t, err)

	// Cleanup after test: remove the DB file
	err = os.Remove(dbPath)
	require.NoError(t, err)
}
