package bankcard

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAddBankCardCommand_Success(t *testing.T) {
	// Capture mock input
	var (
		called       = false
		capturedArgs []string
	)

	mockRunFunc := func(
		ctx context.Context,
		secretName, number, owner, exp, cvv, meta string,
	) error {
		called = true
		capturedArgs = []string{secretName, number, owner, exp, cvv, meta}
		return nil
	}

	root := &cobra.Command{Use: "gophkeeper"}
	RegisterAddBankCardCommand(root, mockRunFunc)

	args := []string{
		"add-bankcard",
		"--secret-name", "testcard",
		"--number", "1234567890123456",
		"--owner", "John Doe",
		"--exp", "12/25",
		"--cvv", "123",
		"--meta", "personal",
	}

	root.SetArgs(args)
	err := root.Execute()

	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, []string{
		"testcard",
		"1234567890123456",
		"John Doe",
		"12/25",
		"123",
		"personal",
	}, capturedArgs)
}

func TestRegisterAddBankCardCommand_MissingRequiredFlags(t *testing.T) {
	mockRunFunc := func(
		ctx context.Context,
		secretName, number, owner, exp, cvv, meta string,
	) error {
		return nil
	}

	root := &cobra.Command{Use: "gophkeeper"}
	RegisterAddBankCardCommand(root, mockRunFunc)

	// Omit required flags
	root.SetArgs([]string{"add-bankcard"})
	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestRunAdd(t *testing.T) {
	const dbFile = "client.db"

	// Clean up DB file before and after test
	os.Remove(dbFile)
	defer os.Remove(dbFile)

	// Open DB and create the table required for bankcard storage
	db, err := sqlx.Connect("sqlite", dbFile)
	require.NoError(t, err)

	// Make sure to create the table your repository expects.
	// Adjust the schema as needed to match your actual bankcard table schema.
	schema := `
CREATE TABLE IF NOT EXISTS bankcard_client (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    secret_name TEXT NOT NULL UNIQUE,
    number TEXT NOT NULL,
    owner TEXT NOT NULL,
    exp TEXT NOT NULL,
    cvv TEXT NOT NULL,
    meta TEXT
);
`
	_, err = db.Exec(schema)
	require.NoError(t, err)
	db.Close()

	ctx := context.Background()
	secretName := "testcard"
	number := "1234567890123456"
	owner := "John Doe"
	exp := "12/25"
	cvv := "123"
	meta := "some meta info"

	// Run the function under test
	err = RunAddBankCard(ctx, secretName, number, owner, exp, cvv, meta)
	require.NoError(t, err)

	// Re-open DB to verify record inserted
	db, err = sqlx.Connect("sqlite", dbFile)
	require.NoError(t, err)
	defer db.Close()

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM bankcard_client WHERE secret_name = ?", secretName)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Verify the inserted record's fields
	var (
		gotNumber string
		gotOwner  string
		gotExp    string
		gotCVV    string
		gotMeta   string
	)
	err = db.QueryRow("SELECT number, owner, exp, cvv, meta FROM bankcard_client WHERE secret_name = ?", secretName).
		Scan(&gotNumber, &gotOwner, &gotExp, &gotCVV, &gotMeta)
	require.NoError(t, err)

	require.Equal(t, number, gotNumber)
	require.Equal(t, owner, gotOwner)
	require.Equal(t, exp, gotExp)
	require.Equal(t, cvv, gotCVV)
	require.Equal(t, meta, gotMeta)
}
