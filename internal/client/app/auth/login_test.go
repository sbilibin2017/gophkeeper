package auth

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRegisterLoginCommand(t *testing.T) {
	// Backup original funcs and restore after test
	origHTTPFunc := loginHTTPFunc
	origGRPCFunc := loginGRPCFunc
	defer func() {
		loginHTTPFunc = origHTTPFunc
		loginGRPCFunc = origGRPCFunc
	}()

	// Mock the HTTP login func
	loginHTTPFunc = func(ctx context.Context, username, password, authURL, tlsCertFile, tlsKeyFile string) (*models.AuthResponse, error) {
		return &models.AuthResponse{Token: "mock-http-login-token"}, nil
	}

	// Mock the gRPC login func
	loginGRPCFunc = func(ctx context.Context, username, password, authURL, tlsCertFile, tlsKeyFile string) (*models.AuthResponse, error) {
		return &models.AuthResponse{Token: "mock-grpc-login-token"}, nil
	}

	root := &cobra.Command{Use: "root"}
	RegisterLoginCommand(root)

	// Test HTTP scheme
	root.SetArgs([]string{
		"login",
		"--username", "alice",
		"--password", "pass",
		"--auth-url", "https://example.com",
		"--tls-client-cert", "cert.pem",
		"--tls-client-key", "key.pem",
	})

	outBuf := new(bytes.Buffer)
	root.SetOut(outBuf)
	root.SetErr(outBuf)

	err := root.Execute()
	require.NoError(t, err)
	output := outBuf.String()
	require.Contains(t, output, "mock-http-login-token")

	// Test gRPC scheme
	root.SetArgs([]string{
		"login",
		"--username", "bob",
		"--password", "pass",
		"--auth-url", "grpc://example.com",
		"--tls-client-cert", "cert.pem",
		"--tls-client-key", "key.pem",
	})
	outBuf.Reset()

	err = root.Execute()
	require.NoError(t, err)
	output = outBuf.String()
	require.Contains(t, output, "mock-grpc-login-token")

	// Test unsupported scheme
	root.SetArgs([]string{
		"login",
		"--username", "charlie",
		"--password", "pass",
		"--auth-url", "ftp://example.com",
		"--tls-client-cert", "cert.pem",
		"--tls-client-key", "key.pem",
	})
	outBuf.Reset()

	err = root.Execute()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "unsupported auth URL scheme"))
}
