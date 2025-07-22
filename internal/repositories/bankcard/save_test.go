package bankcard

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Create table for tests
	createTableQuery := `
	CREATE TABLE bankcard_client (
		secret_name TEXT PRIMARY KEY,
		number TEXT NOT NULL,
		owner TEXT NOT NULL,
		exp TEXT NOT NULL,
		cvv TEXT NOT NULL,
		meta TEXT
	);`
	_, err = db.Exec(createTableQuery)
	require.NoError(t, err)

	return db
}

func TestSaveRepository_Save(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSaveRepository(db)

	// Initial insert
	req := &models.BankCardAddRequest{
		SecretName: "card1",
		Number:     "1234567890123456",
		Owner:      "John Doe",
		Exp:        "12/25",
		CVV:        "123",
		Meta:       nil,
	}
	err := repo.Save(ctx, req)
	require.NoError(t, err)

	// Verify inserted data
	var count int
	err = db.GetContext(ctx, &count, "SELECT COUNT(*) FROM bankcard_client WHERE secret_name = ?", req.SecretName)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Update with different data (simulate change)
	req.Number = "9999888877776666"
	meta := "Updated meta"
	req.Meta = &meta
	err = repo.Save(ctx, req)
	require.NoError(t, err)

	// Verify update
	var updatedNumber string
	var updatedMeta string
	err = db.QueryRowContext(ctx, "SELECT number, meta FROM bankcard_client WHERE secret_name = ?", req.SecretName).
		Scan(&updatedNumber, &updatedMeta)
	require.NoError(t, err)
	require.Equal(t, req.Number, updatedNumber)
	require.Equal(t, meta, updatedMeta)
}
