package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"

	_ "modernc.org/sqlite"
)

func setupSecretTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	schema := `
	CREATE TABLE secrets (
		secret_id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		secret_name TEXT NOT NULL,
		secret_type TEXT NOT NULL,
		encrypted_payload BLOB NOT NULL,
		nonce BLOB NOT NULL,
		meta TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, secret_name)
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestSecretWriteAndReadRepositories(t *testing.T) {
	db := setupSecretTestDB(t)
	defer db.Close()

	ctx := context.Background()
	writeRepo := NewSecretWriteRepository(db)
	readRepo := NewSecretReadRepository(db)

	userID := "user1"
	secretID := "secret1"
	secretName := "email"
	secretType := "password"
	payload := []byte("encrypteddata")
	nonce := []byte("nonce123")
	meta := `{"note":"my secret"}`

	// === Save ===
	secret := &models.SecretDB{
		SecretID:         secretID,
		UserID:           userID,
		SecretName:       secretName,
		SecretType:       secretType,
		EncryptedPayload: payload,
		Nonce:            nonce,
		Meta:             meta,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	err := writeRepo.Save(ctx, secret)
	assert.NoError(t, err)

	// === Get ===
	secretFromDB, err := readRepo.Get(ctx, userID, secretName)
	assert.NoError(t, err)
	assert.Equal(t, secretID, secretFromDB.SecretID)
	assert.Equal(t, userID, secretFromDB.UserID)
	assert.Equal(t, secretName, secretFromDB.SecretName)
	assert.Equal(t, secretType, secretFromDB.SecretType)
	assert.Equal(t, payload, secretFromDB.EncryptedPayload)
	assert.Equal(t, nonce, secretFromDB.Nonce)
	assert.Equal(t, meta, secretFromDB.Meta)

	// === Update ===
	newPayload := []byte("newdata")
	newMeta := `{"note":"updated"}`
	secret.EncryptedPayload = newPayload
	secret.Meta = newMeta
	secret.UpdatedAt = time.Now()

	err = writeRepo.Save(ctx, secret)
	assert.NoError(t, err)

	secretUpdated, err := readRepo.Get(ctx, userID, secretName)
	assert.NoError(t, err)
	assert.Equal(t, newPayload, secretUpdated.EncryptedPayload)
	assert.Equal(t, newMeta, secretUpdated.Meta)

	// === List ===
	secrets, err := readRepo.List(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, secretID, secrets[0].SecretID)
}
