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

func setupTextTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE secret_text (
		secret_name TEXT PRIMARY KEY,
		content TEXT,
		meta TEXT,
		updated_at DATETIME
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestSecretTextClientSaveRepository_Save(t *testing.T) {
	db := setupTextTestDB(t)
	repo := NewSecretTextClientSaveRepository(db)

	type fields struct {
		SecretName string
		Content    string
		Meta       *string
	}

	meta1 := "Initial meta"
	meta2 := "Updated meta"

	tests := []struct {
		name       string
		input      fields
		update     *fields
		verifyFunc func(t *testing.T, got models.SecretTextClient, expected fields)
	}{
		{
			name: "Insert new text secret",
			input: fields{
				SecretName: "text-1",
				Content:    "my very secret note",
				Meta:       &meta1,
			},
			verifyFunc: func(t *testing.T, got models.SecretTextClient, expected fields) {
				assert.Equal(t, expected.SecretName, got.SecretName)
				assert.Equal(t, expected.Content, got.Content)
				if expected.Meta != nil {
					require.NotNil(t, got.Meta)
					assert.Equal(t, *expected.Meta, *got.Meta)
				} else {
					assert.Nil(t, got.Meta)
				}
			},
		},
		{
			name: "Update existing text secret",
			input: fields{
				SecretName: "text-2",
				Content:    "draft text",
				Meta:       nil,
			},
			update: &fields{
				SecretName: "text-2",
				Content:    "final updated text",
				Meta:       &meta2,
			},
			verifyFunc: func(t *testing.T, got models.SecretTextClient, expected fields) {
				assert.Equal(t, expected.SecretName, got.SecretName)
				assert.Equal(t, expected.Content, got.Content)
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

			initial := models.SecretTextClient{
				SecretName: tt.input.SecretName,
				Content:    tt.input.Content,
				Meta:       tt.input.Meta,
				UpdatedAt:  now,
			}

			err := repo.Save(ctx, initial)
			require.NoError(t, err)

			if tt.update != nil {
				updated := models.SecretTextClient{
					SecretName: tt.update.SecretName,
					Content:    tt.update.Content,
					Meta:       tt.update.Meta,
					UpdatedAt:  time.Now().UTC(),
				}
				err := repo.Save(ctx, updated)
				require.NoError(t, err)
			}

			var result models.SecretTextClient
			err = db.Get(&result, `
				SELECT secret_name, content, meta, updated_at
				FROM secret_text
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
