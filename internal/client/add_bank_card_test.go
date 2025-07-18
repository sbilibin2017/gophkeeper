package client

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// setupAddBankCardLocalTestDB creates an in-memory SQLite DB and sets up the required schema.
func setupAddBankCardLocalTestDB2(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open SQLite in-memory DB: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS secret_bank_card_request (
		secret_name TEXT PRIMARY KEY,
		number TEXT NOT NULL,
		owner TEXT NOT NULL,
		exp TEXT NOT NULL,
		cvv TEXT NOT NULL,
		meta TEXT
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestAddBankCardLocal2(t *testing.T) {
	db := setupAddBankCardLocalTestDB2(t)
	defer db.Close()

	ctx := context.Background()

	meta := "personal"
	req := models.BankCardAddRequest{
		SecretName: "card_123",
		Number:     "1234567890123456",
		Owner:      "John Doe",
		Exp:        "12/25",
		CVV:        "123",
		Meta:       &meta,
	}

	// First insert
	err := AddBankCardLocal(ctx, db, req)
	assert.NoError(t, err)

	// Verify insert
	var count int
	err = db.Get(&count, `SELECT COUNT(*) FROM secret_bank_card_request WHERE secret_name = ?`, req.SecretName)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Update (upsert)
	newMeta := "updated-meta"
	req.Number = "1111222233334444"
	req.Owner = "Jane Smith"
	req.Meta = &newMeta

	err = AddBankCardLocal(ctx, db, req)
	assert.NoError(t, err)

	// Verify updated values
	var (
		secretName, number, owner, exp, cvv string
		metaVal                             *string
	)
	err = db.QueryRowx(`
		SELECT secret_name, number, owner, exp, cvv, meta
		FROM secret_bank_card_request
		WHERE secret_name = ?
	`, req.SecretName).Scan(&secretName, &number, &owner, &exp, &cvv, &metaVal)
	assert.NoError(t, err)

	assert.Equal(t, req.SecretName, secretName)
	assert.Equal(t, req.Number, number)
	assert.Equal(t, req.Owner, owner)
	assert.Equal(t, req.Exp, exp)
	assert.Equal(t, req.CVV, cvv)
	assert.NotNil(t, metaVal)
	assert.Equal(t, *req.Meta, *metaVal)
}
