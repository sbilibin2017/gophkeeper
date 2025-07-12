package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestMigrate_Success(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbConn.Close()

	// Ожидаем 4 CREATE TABLE IF NOT EXISTS запросов
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS username_passwords").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS texts").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS binaries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS bank_cards").WillReturnResult(sqlmock.NewResult(1, 1))

	err = Migrate(dbConn)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMigrate_Error(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbConn.Close()

	// Возвращаем ошибку при первом запросе
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS username_passwords").
		WillReturnError(sqlmock.ErrCancelled)

	err = Migrate(dbConn)
	require.Error(t, err)
	require.Equal(t, sqlmock.ErrCancelled, err)
}
