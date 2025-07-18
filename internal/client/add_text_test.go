package client

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// setupAddTextLocalTestDB creates an in-memory SQLite DB and sets up the required schema.
func setupAddTextLocalTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open SQLite in-memory DB: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS secret_text_request (
		secret_name TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		meta TEXT
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestAddTextLocal(t *testing.T) {
	db := setupAddTextLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()

	meta := "first meta"
	req := models.TextAddRequest{
		SecretName: "text_001",
		Content:    "initial content",
		Meta:       &meta,
	}

	// First insert
	err := AddTextLocal(ctx, db, req)
	assert.NoError(t, err)

	// Verify insert
	var count int
	err = db.Get(&count, `SELECT COUNT(*) FROM secret_text_request WHERE secret_name = ?`, req.SecretName)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Update (upsert)
	newMeta := "updated meta"
	req.Content = "updated content"
	req.Meta = &newMeta

	err = AddTextLocal(ctx, db, req)
	assert.NoError(t, err)

	// Verify updated values
	var (
		secretName string
		content    string
		metaVal    *string
	)
	err = db.QueryRowx(`
		SELECT secret_name, content, meta
		FROM secret_text_request
		WHERE secret_name = ?
	`, req.SecretName).Scan(&secretName, &content, &metaVal)
	assert.NoError(t, err)

	assert.Equal(t, req.SecretName, secretName)
	assert.Equal(t, req.Content, content)
	assert.NotNil(t, metaVal)
	assert.Equal(t, *req.Meta, *metaVal)
}
