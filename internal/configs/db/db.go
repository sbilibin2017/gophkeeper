package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// NewDB устанавливает соединение с базой данных и возвращает sqlx.DB.
func NewDB(driver string, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// RunMigrations применяет миграции из указанной директории, используя goose.
func RunMigrations(db *sqlx.DB, driver string, pathToDir string) error {
	if err := goose.SetDialect(driver); err != nil {
		return err
	}
	if err := goose.Up(db.DB, pathToDir); err != nil {
		return err
	}
	return nil
}
