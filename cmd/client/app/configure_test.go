package app

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestConfigureCommand_TokenProvided(t *testing.T) {
	cmd := appCommand()

	// Set the token flag
	cmd.SetArgs([]string{"configure", "--token", "my_test_token"})

	err := cmd.Execute()
	assert.NoError(t, err, "command should execute without error")

	token := os.Getenv("GOPHKEEPER_TOKEN")
	assert.Equal(t, "my_test_token", token, "GOPHKEEPER_TOKEN should match provided token")
}

func TestConfigureCommand_TokenMissing(t *testing.T) {
	cmd := appCommand()

	// No token flag provided
	cmd.SetArgs([]string{"configure"})

	err := cmd.Execute()
	assert.Error(t, err, "command should return error if token is not provided")
	assert.Contains(t, err.Error(), "token is required")
}

// appCommand creates a root command with `configure` subcommand
func appCommand() *cobra.Command {
	root := &cobra.Command{Use: "client"}
	root.AddCommand(newConfigureCommand())
	return root
}
