package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	rootCmd := NewRootCommand()

	assert.NotNil(t, rootCmd, "root command should not be nil")
	assert.Equal(t, "gophkeeper", rootCmd.Use, "command use should be 'gophkeeper'")
	assert.Contains(t, rootCmd.Short, "password manager", "short description should mention password manager")
	assert.Contains(t, rootCmd.Long, "client-server system", "long description should mention client-server system")

	// Check some expected subcommands are registered
	subCommands := rootCmd.Commands()
	var subCommandNames []string
	for _, cmd := range subCommands {
		subCommandNames = append(subCommandNames, cmd.Name())
	}

	expectedCommands := []string{
		"register",
		"login",
		"add-bank-card",
		"add-binary-secret",
		"add-text-secret",
		"add-username-password",
		"get-secret",
		"list-secrets",
		"delete-secret",
		"sync",
	}

	for _, expected := range expectedCommands {
		assert.Contains(t, subCommandNames, expected, "should contain subcommand "+expected)
	}
}
