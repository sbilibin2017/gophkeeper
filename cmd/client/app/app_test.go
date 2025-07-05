package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClientCommand(t *testing.T) {
	cmd := NewAppCommand()

	require.NotNil(t, cmd)
	require.Equal(t, "gophkeeper", cmd.Use)
	require.Equal(t, "GophKeeper is a secure personal data manager CLI", cmd.Short)

	// Collect subcommand names
	subCmds := cmd.Commands()
	require.NotEmpty(t, subCmds)

	names := make(map[string]struct{}, len(subCmds))
	for _, c := range subCmds {
		names[c.Name()] = struct{}{}
	}

	// Check expected subcommands are present
	require.Contains(t, names, "build-info")
	require.Contains(t, names, "usage")
	require.Contains(t, names, "register")
}
