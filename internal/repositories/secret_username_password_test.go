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

func setupUsernamePasswordTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_username_password (
		secret_name TEXT PRIMARY KEY,
		username TEXT,
		password TEXT,
		meta TEXT,
		updated_at DATETIME
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSecretUsernamePasswordClientSaveRepository_Save(t *testing.T) {
	db := setupUsernamePasswordTestDB(t)
	repo := NewSecretUsernamePasswordClientSaveRepository(db)

	type fields struct {
		SecretName string
		Username   string
		Password   string
		Meta       *string
	}

	meta1 := "initial meta"
	meta2 := "updated meta"

	tests := []struct {
		name       string
		input      fields
		update     *fields
		verifyFunc func(t *testing.T, got models.SecretUsernamePasswordClient, expected fields)
	}{
		{
			name: "Insert new secret",
			input: fields{
				SecretName: "secret1",
				Username:   "user1",
				Password:   "pass1",
				Meta:       &meta1,
			},
			verifyFunc: func(t *testing.T, got models.SecretUsernamePasswordClient, expected fields) {
				assert.Equal(t, expected.SecretName, got.SecretName)
				assert.Equal(t, expected.Username, got.Username)
				assert.Equal(t, expected.Password, got.Password)
				if expected.Meta != nil {
					require.NotNil(t, got.Meta)
					assert.Equal(t, *expected.Meta, *got.Meta)
				} else {
					assert.Nil(t, got.Meta)
				}
			},
		},
		{
			name: "Update existing secret",
			input: fields{
				SecretName: "secret2",
				Username:   "user2",
				Password:   "pass2",
				Meta:       nil,
			},
			update: &fields{
				SecretName: "secret2",
				Username:   "user2_updated",
				Password:   "pass2_updated",
				Meta:       &meta2,
			},
			verifyFunc: func(t *testing.T, got models.SecretUsernamePasswordClient, expected fields) {
				assert.Equal(t, expected.SecretName, got.SecretName)
				assert.Equal(t, expected.Username, got.Username)
				assert.Equal(t, expected.Password, got.Password)
				if expected.Meta != nil {
					require.NotNil(t, got.Meta)
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
			now := time.Now().UTC()

			initial := models.SecretUsernamePasswordClient{
				SecretName: tt.input.SecretName,
				Username:   tt.input.Username,
				Password:   tt.input.Password,
				Meta:       tt.input.Meta,
				UpdatedAt:  now,
			}

			err := repo.Save(ctx, initial)
			require.NoError(t, err)

			if tt.update != nil {
				updated := models.SecretUsernamePasswordClient{
					SecretName: tt.update.SecretName,
					Username:   tt.update.Username,
					Password:   tt.update.Password,
					Meta:       tt.update.Meta,
					UpdatedAt:  time.Now().UTC(),
				}
				err := repo.Save(ctx, updated)
				require.NoError(t, err)
			}

			var result models.SecretUsernamePasswordClient
			err = db.Get(&result, `
				SELECT secret_name, username, password, meta, updated_at
				FROM secret_username_password
				WHERE secret_name = ?`, tt.input.SecretName,
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
