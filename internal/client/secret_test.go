package client

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func setupSecretTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Create all required tables
	require.NoError(t, CreateBankCardRequestTable(db))
	require.NoError(t, CreateUsernamePasswordRequestTable(db))
	require.NoError(t, CreateTextRequestTable(db))
	require.NoError(t, CreateBinaryRequestTable(db))

	return db
}

func TestAddBankCardSecret(t *testing.T) {
	ctx := context.Background()
	db := setupSecretTestDB(t)
	defer db.Close()

	req := models.BankCardAddRequest{
		SecretName: "card1",
		Number:     "1234567890123456",
		Owner:      "John Doe",
		Exp:        "12/24",
		CVV:        "123",
		Meta:       nil,
	}

	err := AddBankCardSecret(ctx, db, req)
	require.NoError(t, err)
}

func TestAddUsernamePasswordSecret(t *testing.T) {
	ctx := context.Background()
	db := setupSecretTestDB(t)
	defer db.Close()

	req := models.UsernamePasswordAddRequest{
		SecretName: "user1",
		Username:   "johndoe",
		Password:   "s3cr3t",
		Meta:       nil,
	}

	err := AddUsernamePasswordSecret(ctx, db, req)
	require.NoError(t, err)
}

func TestAddTextSecret(t *testing.T) {
	ctx := context.Background()
	db := setupSecretTestDB(t)
	defer db.Close()

	req := models.TextAddRequest{
		SecretName: "text1",
		Content:    "my secret text",
		Meta:       nil,
	}

	err := AddTextSecret(ctx, db, req)
	require.NoError(t, err)
}

func TestAddBinarySecret(t *testing.T) {
	ctx := context.Background()
	db := setupSecretTestDB(t)
	defer db.Close()

	req := models.AddSecretBinaryRequest{
		SecretName: "binary1",
		Data:       []byte{0x01, 0x02, 0x03},
		Meta:       nil,
	}

	err := AddBinarySecret(ctx, db, req)
	require.NoError(t, err)
}
