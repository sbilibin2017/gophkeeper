package db

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// Opt is a functional option type for configuring a DB connection.
type Opt func(*sqlx.DB)

// NewDB creates and configures a database connection using provided options.
func NewDB(driver string, dsn string, opts ...Opt) (*sqlx.DB, error) {
	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Apply options
	for _, opt := range opts {
		opt(db)
	}

	return db, nil
}

// WithMaxOpenConns sets the maximum number of open connections.
func WithMaxOpenConns(n int) Opt {
	return func(db *sqlx.DB) {
		db.SetMaxOpenConns(n)
	}
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) Opt {
	return func(db *sqlx.DB) {
		db.SetMaxIdleConns(n)
	}
}

// WithConnMaxLifetime sets the maximum lifetime of a connection.
func WithConnMaxLifetime(d time.Duration) Opt {
	return func(db *sqlx.DB) {
		db.SetConnMaxLifetime(d)
	}
}

// RunMigrations runs Goose migrations against the provided database.
func RunMigrations(db *sqlx.DB, driver string, pathToDir string) error {
	if err := goose.SetDialect(driver); err != nil {
		return err
	}
	if err := goose.Up(db.DB, pathToDir); err != nil {
		return err
	}
	return nil
}
