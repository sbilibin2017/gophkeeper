package db

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// NewDB opens a connection to an SQLite database from the specified file path
// and verifies the connection.
func NewDB(pathToDB string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", pathToDB)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
