package client

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// setupAddBinaryLocalTestDB creates an in-memory SQLite DB and sets up the required schema.
func setupAddBinaryLocalTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open SQLite in-memory DB: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS secret_binary_request (
		secret_name TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		meta TEXT
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestAddBinaryLocal(t *testing.T) {
	db := setupAddBinaryLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()

	meta := "initial meta"
	req := models.BinaryAddRequest{
		SecretName: "binary_001",
		Data:       []byte{0x01, 0x02, 0x03, 0x04},
		Meta:       &meta,
	}

	// First insert
	err := AddBinaryLocal(ctx, db, req)
	assert.NoError(t, err)

	// Verify insert
	var count int
	err = db.Get(&count, `SELECT COUNT(*) FROM secret_binary_request WHERE secret_name = ?`, req.SecretName)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Update (upsert)
	newMeta := "updated meta"
	req.Data = []byte{0xFF, 0xEE, 0xDD, 0xCC}
	req.Meta = &newMeta

	err = AddBinaryLocal(ctx, db, req)
	assert.NoError(t, err)

	// Verify updated values
	var (
		secretName string
		data       []byte
		metaVal    *string
	)
	err = db.QueryRowx(`
		SELECT secret_name, data, meta
		FROM secret_binary_request
		WHERE secret_name = ?
	`, req.SecretName).Scan(&secretName, &data, &metaVal)
	assert.NoError(t, err)

	assert.Equal(t, req.SecretName, secretName)
	assert.Equal(t, req.Data, data)
	assert.NotNil(t, metaVal)
	assert.Equal(t, *req.Meta, *metaVal)
}
