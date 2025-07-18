package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// NewDB establishes a connection to the database and returns a sqlx.DB instance.
// Parameters:
// - driver: the database driver name (e.g., "sqlite").
// - dsn: the data source name or connection string.
// Returns:
// - a pointer to sqlx.DB if successful.
// - an error if connection fails.
func NewDB(driver string, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// RunMigrations applies database migrations from the specified directory using goose.
// Parameters:
// - db: the sqlx.DB instance connected to the database.
// - driver: the database driver name (e.g., "sqlite").
// - pathToDir: the directory path containing migration files.
// Returns:
// - an error if migration setup or execution fails.
func RunMigrations(db *sqlx.DB, driver string, pathToDir string) error {
	if err := goose.SetDialect(driver); err != nil {
		return err
	}
	if err := goose.Up(db.DB, pathToDir); err != nil {
		return err
	}
	return nil
}
