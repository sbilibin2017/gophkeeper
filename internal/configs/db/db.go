package db

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func NewDB() (*sqlx.DB, error) {
	// Get the absolute path of this source file (db.go)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get runtime caller info")
	}

	// Get the directory of this source file (should be .../internal/configs/db)
	dir := filepath.Dir(filename)

	// Compose absolute path to db.sqlite inside the same directory
	dbPath := filepath.Join(dir, "db.sqlite")

	// Ensure the directory exists (should already exist, but just in case)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
	}

	// Create file if not exists (optional)
	f, err := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create database file %s: %w", dbPath, err)
	}
	f.Close()

	// Open the SQLite database using absolute path
	db, err := sqlx.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database at %s: %w", dbPath, err)
	}

	// Ping to verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	return db, nil
}
