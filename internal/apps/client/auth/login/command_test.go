package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRegisterCommand_RunE(t *testing.T) {
	runHTTPFunc := func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
		if username == "httpuser" && password == "httppass" {
			return &models.AuthResponse{Token: "http_token"}, nil
		}
		return nil, errors.New("http login failed")
	}

	runGRPCFunc := func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
		if username == "grpcuser" && password == "grpcpass" {
			return &models.AuthResponse{Token: "grpc_token"}, nil
		}
		return nil, errors.New("grpc login failed")
	}

	tests := []struct {
		name        string
		args        []string
		wantOutput  string
		wantErrPart string
	}{
		{
			name:       "Successful gRPC login",
			args:       []string{"login", "--username", "grpcuser", "--password", "grpcpass", "--auth-url", "grpc://localhost", "--tls-client-cert", "cert.pem", "--tls-client-key", "key.pem"},
			wantOutput: "grpc_token\n",
		},
		{
			name:       "Successful HTTP login",
			args:       []string{"login", "--username", "httpuser", "--password", "httppass", "--auth-url", "https://example.com", "--tls-client-cert", "cert.pem", "--tls-client-key", "key.pem"},
			wantOutput: "http_token\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &cobra.Command{Use: "root"}
			RegisterCommand(root, runHTTPFunc, runGRPCFunc)

			var output bytes.Buffer
			root.SetOut(&output)
			root.SetErr(&output)
			root.SetArgs(tt.args)

			err := root.Execute()

			if tt.wantErrPart != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				assert.Empty(t, output.String())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOutput, output.String())
			}
		})
	}
}
