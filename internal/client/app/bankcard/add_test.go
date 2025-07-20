package bankcard

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAddBankCardCommand_Success(t *testing.T) {
	// Backup original to restore later
	original := addBankCardClient
	defer func() { addBankCardClient = original }()

	// Capture mock input
	var (
		called       = false
		capturedArgs []string
	)

	addBankCardClient = func(
		ctx context.Context,
		secretName, number, owner, exp, cvv, meta string,
	) error {
		called = true
		capturedArgs = []string{secretName, number, owner, exp, cvv, meta}
		return nil
	}

	root := &cobra.Command{Use: "gophkeeper"}
	RegisterAddCommand(root)

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
	original := addBankCardClient
	defer func() { addBankCardClient = original }()

	root := &cobra.Command{Use: "gophkeeper"}
	RegisterAddCommand(root)

	// Omit required flags
	root.SetArgs([]string{"add-bankcard"})
	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}
