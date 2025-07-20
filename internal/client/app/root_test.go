package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	require.NotNil(t, cmd, "NewRootCommand() should not return nil")
	require.Equal(t, "gophkeeper", cmd.Use, "command Use field mismatch")
	require.Contains(t, cmd.Short, "password manager", "command Short description mismatch")
	require.Contains(t, cmd.Long, "GophKeeper is a client-server system", "command Long description mismatch")
}
