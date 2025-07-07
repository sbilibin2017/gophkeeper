package app

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestNewAppCommand(t *testing.T) {
	cmd := NewAppCommand()
	require.NotNil(t, cmd)

	// Check main command fields
	require.Equal(t, "gophkeeper", cmd.Use)
	require.Contains(t, cmd.Short, "CLI tool")
	require.Contains(t, cmd.Long, "Available commands:")

	// Check child commands exist
	subCmds := cmd.Commands()
	require.NotEmpty(t, subCmds)

	expectedCmds := []string{
		"build-info",
		"register",
		"login",
		"add",
		"get",
		"list",
		"sync",
	}

	for _, name := range expectedCmds {
		require.Truef(t, containsCommand(subCmds, name), "expected command %q", name)
	}
}

// containsCommand checks if a command with the given name exists in cmds
func containsCommand(cmds []*cobra.Command, name string) bool {
	for _, c := range cmds {
		if c.Name() == name {
			return true
		}
	}
	return false
}
