package client

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRegisterLogoutCommand(t *testing.T) {
	// Backup original funcs and restore after test
	origHTTPFunc := logoutHTTPFunc
	origGRPCFunc := logoutGRPCFunc
	defer func() {
		logoutHTTPFunc = origHTTPFunc
		logoutGRPCFunc = origGRPCFunc
	}()

	// Mock HTTP logout function
	logoutHTTPFunc = func(ctx context.Context, token, authURL, tlsCertFile, tlsKeyFile string) error {
		return nil // simulate success
	}

	// Mock gRPC logout function
	logoutGRPCFunc = func(ctx context.Context, token, authURL, tlsCertFile, tlsKeyFile string) error {
		return nil // simulate success
	}

	root := &cobra.Command{Use: "root"}
	RegisterLogoutCommand(root)

	// Test HTTP logout
	root.SetArgs([]string{
		"logout",
		"--token", "dummy-token",
		"--auth-url", "https://example.com",
		"--tls-client-cert", "cert.pem",
		"--tls-client-key", "key.pem",
	})

	outBuf := new(bytes.Buffer)
	root.SetOut(outBuf)
	root.SetErr(outBuf)

	err := root.Execute()
	require.NoError(t, err)

	// Test gRPC logout
	root.SetArgs([]string{
		"logout",
		"--token", "dummy-token",
		"--auth-url", "grpc://example.com",
		"--tls-client-cert", "cert.pem",
		"--tls-client-key", "key.pem",
	})
	outBuf.Reset()

	err = root.Execute()
	require.NoError(t, err)

	// Test unsupported URL scheme error
	root.SetArgs([]string{
		"logout",
		"--token", "dummy-token",
		"--auth-url", "ftp://example.com",
		"--tls-client-cert", "cert.pem",
		"--tls-client-key", "key.pem",
	})
	outBuf.Reset()

	err = root.Execute()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "unsupported auth URL scheme"))
}
