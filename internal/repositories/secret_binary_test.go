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

func setupBinaryTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_binary (
		secret_name TEXT PRIMARY KEY,
		data BLOB,
		meta TEXT,
		updated_at DATETIME
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSecretBinaryClientSaveRepository_Save(t *testing.T) {
	db := setupBinaryTestDB(t)
	repo := NewSecretBinaryClientSaveRepository(db)

	type fields struct {
		SecretName string  `json:"secret_name" db:"secret_name"`
		Data       []byte  `json:"data" db:"data"`
		Meta       *string `json:"meta,omitempty" db:"meta"`
	}

	meta1 := "Initial binary meta"
	meta2 := "Updated binary meta"

	tests := []struct {
		name       string
		input      fields
		update     *fields
		verifyFunc func(t *testing.T, got models.SecretBinaryClient, expected fields)
	}{
		{
			name: "Insert binary secret",
			input: fields{
				SecretName: "bin-1",
				Data:       []byte{0x01, 0x02, 0x03},
				Meta:       &meta1,
			},
			verifyFunc: func(t *testing.T, got models.SecretBinaryClient, expected fields) {
				assert.Equal(t, expected.SecretName, got.SecretName)
				assert.Equal(t, expected.Data, got.Data)
				if expected.Meta != nil {
					require.NotNil(t, got.Meta)
					assert.Equal(t, *expected.Meta, *got.Meta)
				} else {
					assert.Nil(t, got.Meta)
				}
			},
		},
		{
			name: "Update existing binary secret",
			input: fields{
				SecretName: "bin-2",
				Data:       []byte("original-data"),
				Meta:       nil,
			},
			update: &fields{
				SecretName: "bin-2",
				Data:       []byte("new-updated-data"),
				Meta:       &meta2,
			},
			verifyFunc: func(t *testing.T, got models.SecretBinaryClient, expected fields) {
				assert.Equal(t, expected.SecretName, got.SecretName)
				assert.Equal(t, expected.Data, got.Data)
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

			initial := models.SecretBinaryClient{
				SecretName: tt.input.SecretName,
				Data:       tt.input.Data,
				Meta:       tt.input.Meta,
				UpdatedAt:  now,
			}

			err := repo.Save(ctx, initial)
			require.NoError(t, err)

			if tt.update != nil {
				updated := models.SecretBinaryClient{
					SecretName: tt.update.SecretName,
					Data:       tt.update.Data,
					Meta:       tt.update.Meta,
					UpdatedAt:  time.Now().UTC(),
				}
				err := repo.Save(ctx, updated)
				require.NoError(t, err)
			}

			var result models.SecretBinaryClient
			err = db.Get(&result, `
				SELECT secret_name, data, meta, updated_at
				FROM secret_binary
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
