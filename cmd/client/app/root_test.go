package app_test

import (
	"testing"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	cmd := app.NewCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "gophkeeper", cmd.Use)
	assert.Contains(t, cmd.Short, "GophKeeper")

	expectedSubcommands := []string{
		"auth",
		"add-username-password",
		"add-text",
		"add-binary",
		"add-bank-card",
	}

	subCmds := cmd.Commands()
	found := make(map[string]bool)

	for _, c := range subCmds {
		found[c.Name()] = true
	}

	for _, name := range expectedSubcommands {
		assert.True(t, found[name], "expected to find '%s' subcommand", name)
	}
}
