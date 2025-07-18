package client

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

// setupAddUsernamePasswordLocalTestDB creates an in-memory SQLite DB and sets up the required schema.
func setupAddUsernamePasswordLocalTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open SQLite in-memory DB: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS secret_username_password_request (
		secret_name TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		meta TEXT
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func TestAddUsernamePasswordLocal(t *testing.T) {
	db := setupAddUsernamePasswordLocalTestDB(t)
	defer db.Close()

	ctx := context.Background()

	meta := "initial meta"
	req := models.UsernamePasswordAddRequest{
		SecretName: "login_001",
		Username:   "user1",
		Password:   "pass123",
		Meta:       &meta,
	}

	// First insert
	err := AddUsernamePasswordLocal(ctx, db, req)
	assert.NoError(t, err)

	// Verify insert
	var count int
	err = db.Get(&count, `SELECT COUNT(*) FROM secret_username_password_request WHERE secret_name = ?`, req.SecretName)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Update (upsert)
	newMeta := "updated meta"
	req.Username = "user2"
	req.Password = "newpass456"
	req.Meta = &newMeta

	err = AddUsernamePasswordLocal(ctx, db, req)
	assert.NoError(t, err)

	// Verify updated values
	var (
		secretName string
		username   string
		password   string
		metaVal    *string
	)
	err = db.QueryRowx(`
		SELECT secret_name, username, password, meta
		FROM secret_username_password_request
		WHERE secret_name = ?
	`, req.SecretName).Scan(&secretName, &username, &password, &metaVal)
	assert.NoError(t, err)

	assert.Equal(t, req.SecretName, secretName)
	assert.Equal(t, req.Username, username)
	assert.Equal(t, req.Password, password)
	assert.NotNil(t, metaVal)
	assert.Equal(t, *req.Meta, *metaVal)
}
