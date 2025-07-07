package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLoginCommand_RequiredFlags(t *testing.T) {
	cmd := newLoginCommand()

	// No args should error for missing required flags
	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "required flag(s) \"password\", \"server-url\", \"username\" not set")
}

func TestNewLoginCommand_UnsupportedScheme(t *testing.T) {
	cmd := newLoginCommand()
	cmd.SetArgs([]string{
		"--server-url", "ftp://invalid",
		"--username", "user",
		"--password", "pass",
	})

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported URL scheme")
}

func TestNewLoginCommand_ValidHTTPURL_ButFails(t *testing.T) {
	cmd := newLoginCommand()
	cmd.SetArgs([]string{
		"--server-url", "http://localhost",
		"--username", "user",
		"--password", "pass",
	})

	err := cmd.Execute()
	// This will likely fail unless you have a real server running on http://localhost
	require.Error(t, err)
}

func TestNewLoginCommand_ValidGRPCURL_ButFails(t *testing.T) {
	cmd := newLoginCommand()
	cmd.SetArgs([]string{
		"--server-url", "grpc://localhost:12345",
		"--username", "user",
		"--password", "pass",
	})

	err := cmd.Execute()
	// This will likely fail unless you have a grpc server running on localhost:12345
	require.Error(t, err)
}
