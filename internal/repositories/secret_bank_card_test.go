package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_bank_card (
		secret_name TEXT NOT NULL,
		owner TEXT NOT NULL,
		number TEXT,
		exp TEXT,
		cvv TEXT,
		meta TEXT,
		updated_at DATETIME,
		PRIMARY KEY (secret_name, owner)
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSecretBankCardClientSaveRepository_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSecretBankCardClientSaveRepository(db)

	type fields struct {
		SecretName string
		Owner      string
		Number     string
		Exp        string
		CVV        string
		Meta       *string
	}
	now := time.Now().UTC()

	meta1 := "Initial meta"
	meta2 := "Updated meta"

	tests := []struct {
		name       string
		input      fields
		update     *fields // если nil — значит это insert
		verifyFunc func(t *testing.T, got models.SecretBankCardClient, expected fields)
	}{
		{
			name: "Insert new secret",
			input: fields{
				SecretName: "card-1",
				Owner:      "user1",
				Number:     "4111111111111111",
				Exp:        "12/25",
				CVV:        "123",
				Meta:       &meta1,
			},
			verifyFunc: func(t *testing.T, got models.SecretBankCardClient, expected fields) {
				assert.Equal(t, expected.Number, got.Number)
				assert.Equal(t, expected.CVV, got.CVV)
				assert.Equal(t, expected.Exp, got.Exp)
				assert.Equal(t, expected.Owner, got.Owner)
				assert.Equal(t, expected.SecretName, got.SecretName)
				if expected.Meta != nil {
					assert.NotNil(t, got.Meta)
					assert.Equal(t, *expected.Meta, *got.Meta)
				} else {
					assert.Nil(t, got.Meta)
				}
			},
		},
		{
			name: "Update existing secret",
			input: fields{
				SecretName: "card-2",
				Owner:      "user2",
				Number:     "4000000000000000",
				Exp:        "01/26",
				CVV:        "999",
				Meta:       nil,
			},
			update: &fields{
				SecretName: "card-2",
				Owner:      "user2",
				Number:     "5555555555554444",
				Exp:        "02/27",
				CVV:        "888",
				Meta:       &meta2,
			},
			verifyFunc: func(t *testing.T, got models.SecretBankCardClient, expected fields) {
				assert.Equal(t, expected.Number, got.Number)
				assert.Equal(t, expected.CVV, got.CVV)
				assert.Equal(t, expected.Exp, got.Exp)
				assert.Equal(t, expected.Owner, got.Owner)
				assert.Equal(t, expected.SecretName, got.SecretName)
				if expected.Meta != nil {
					assert.NotNil(t, got.Meta)
					assert.Equal(t, *expected.Meta, *got.Meta)
				} else {
					assert.Nil(t, got.Meta)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			initial := models.SecretBankCardClient{
				SecretName: tt.input.SecretName,
				Owner:      tt.input.Owner,
				Number:     tt.input.Number,
				Exp:        tt.input.Exp,
				CVV:        tt.input.CVV,
				Meta:       tt.input.Meta,
				UpdatedAt:  now,
			}

			err := repo.Save(ctx, initial)
			require.NoError(t, err)

			// если есть update — применяем и повторно сохраняем
			if tt.update != nil {
				updated := models.SecretBankCardClient{
					SecretName: tt.update.SecretName,
					Owner:      tt.update.Owner,
					Number:     tt.update.Number,
					Exp:        tt.update.Exp,
					CVV:        tt.update.CVV,
					Meta:       tt.update.Meta,
					UpdatedAt:  time.Now().UTC(),
				}
				err := repo.Save(ctx, updated)
				require.NoError(t, err)
			}

			var result models.SecretBankCardClient
			err = db.Get(&result, `
				SELECT secret_name, owner, number, exp, cvv, meta, updated_at
				FROM secret_bank_card
				WHERE secret_name = ? AND owner = ?`,
				tt.input.SecretName, tt.input.Owner,
			)
			require.NoError(t, err)

			expected := tt.input
			if tt.update != nil {
				expected = *tt.update
			}
			tt.verifyFunc(t, result, expected)
		})
	}
}
