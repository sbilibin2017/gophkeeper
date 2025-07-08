package services

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Создадим необходимые таблицы (по структуре твоих функций)
	schema := `
	CREATE TABLE login_passwords (
		secret_id TEXT PRIMARY KEY,
		login TEXT,
		password TEXT,
		meta TEXT
	);
	CREATE TABLE texts (
		secret_id TEXT PRIMARY KEY,
		content TEXT,
		meta TEXT,
		updated_at DATETIME
	);
	CREATE TABLE cards (
		secret_id TEXT PRIMARY KEY,
		number TEXT,
		holder TEXT,
		exp_month INTEGER,
		exp_year INTEGER,
		cvv TEXT,
		meta TEXT,
		updated_at DATETIME
	);
	CREATE TABLE binaries (
		secret_id TEXT PRIMARY KEY,
		data TEXT,
		meta TEXT,
		updated_at DATETIME
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestAddLoginPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	secret := &models.LoginPassword{
		SecretID: "secret1",
		Login:    "user1",
		Password: "pass1",
		Meta:     map[string]string{"key1": "val1"},
	}

	err := AddLoginPassword(context.Background(), db, secret)
	require.NoError(t, err)

	var login, password, meta string
	err = db.QueryRow("SELECT login, password, meta FROM login_passwords WHERE secret_id = ?", secret.SecretID).
		Scan(&login, &password, &meta)
	require.NoError(t, err)

	assert.Equal(t, secret.Login, login)
	assert.Equal(t, secret.Password, password)
	assert.Contains(t, meta, `"key1":"val1"`)
}

func TestAddText(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now().UTC()
	secret := &models.Text{
		SecretID:  "text1",
		Content:   "some text content",
		Meta:      map[string]string{"metaKey": "metaValue"},
		UpdatedAt: now,
	}

	err := AddText(context.Background(), db, secret)
	require.NoError(t, err)

	var content, meta string
	var updatedAt time.Time
	err = db.QueryRow("SELECT content, meta, updated_at FROM texts WHERE secret_id = ?", secret.SecretID).
		Scan(&content, &meta, &updatedAt)
	require.NoError(t, err)

	assert.Equal(t, secret.Content, content)
	assert.Contains(t, meta, `"metaKey":"metaValue"`)
	assert.WithinDuration(t, now, updatedAt, time.Second)
}

func TestAddCard(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now().UTC()
	secret := &models.Card{
		SecretID:  "card1",
		Number:    "1234123412341234",
		Holder:    "John Doe",
		ExpMonth:  12,
		ExpYear:   2030,
		CVV:       "123",
		Meta:      map[string]string{"type": "visa"},
		UpdatedAt: now,
	}

	err := AddCard(context.Background(), db, secret)
	require.NoError(t, err)

	var number, holder, cvv, meta string
	var expMonth, expYear int
	var updatedAt time.Time

	err = db.QueryRow(`SELECT number, holder, exp_month, exp_year, cvv, meta, updated_at 
		FROM cards WHERE secret_id = ?`, secret.SecretID).
		Scan(&number, &holder, &expMonth, &expYear, &cvv, &meta, &updatedAt)
	require.NoError(t, err)

	assert.Equal(t, secret.Number, number)
	assert.Equal(t, secret.Holder, holder)
	assert.Equal(t, secret.ExpMonth, expMonth)
	assert.Equal(t, secret.ExpYear, expYear)
	assert.Equal(t, secret.CVV, cvv)
	assert.Contains(t, meta, `"type":"visa"`)
	assert.WithinDuration(t, now, updatedAt, time.Second)
}

func TestAddBinary(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now().UTC()
	data := []byte{0x01, 0x02, 0x03}
	secret := &models.Binary{
		SecretID:  "bin1",
		Data:      data,
		Meta:      map[string]string{"format": "bin"},
		UpdatedAt: now,
	}

	err := AddBinary(context.Background(), db, secret)
	require.NoError(t, err)

	var base64Data, meta string
	var updatedAt time.Time

	err = db.QueryRow("SELECT data, meta, updated_at FROM binaries WHERE secret_id = ?", secret.SecretID).
		Scan(&base64Data, &meta, &updatedAt)
	require.NoError(t, err)

	expectedBase64 := "AQID" // base64 of 0x01 0x02 0x03
	assert.Equal(t, expectedBase64, base64Data)
	assert.Contains(t, meta, `"format":"bin"`)
	assert.WithinDuration(t, now, updatedAt, time.Second)
}
