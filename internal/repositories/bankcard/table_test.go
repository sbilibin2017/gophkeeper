package bankcard

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func setupDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)
	return db
}

func TestCreateClientTable_Success(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)
	defer db.Close()

	err := CreateClientTable(ctx, db)
	require.NoError(t, err)

	// Try creating again, expect error because table exists
	err = CreateClientTable(ctx, db)
	assert.Error(t, err)
}

func TestDropClientTable_Success(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)
	defer db.Close()

	// Drop table before creation - expect error (table doesn't exist)
	err := DropClientTable(ctx, db)
	assert.Error(t, err)

	// Create table first
	err = CreateClientTable(ctx, db)
	require.NoError(t, err)

	// Drop table now - should succeed
	err = DropClientTable(ctx, db)
	require.NoError(t, err)

	// Drop again - should error (table dropped already)
	err = DropClientTable(ctx, db)
	assert.Error(t, err)
}
