package client_test

import (
	"bufio"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func TestResolveConflictUsernamePassword(t *testing.T) {
	ctx := context.Background()

	server := &models.UsernamePasswordResponse{
		SecretName: "serverName",
		Username:   "serverUser",
		Password:   "serverPass",
	}

	clientReq := &models.UsernamePasswordAddRequest{
		SecretName: "clientName",
		Username:   "clientUser",
		Password:   "clientPass",
	}

	tests := []struct {
		name            string
		resolveStrategy string
		input           string
		expectedOutput  *models.UsernamePasswordAddRequest
		expectError     bool
	}{
		{
			name:            "ResolveStrategyServer returns nil",
			resolveStrategy: models.ResolveStrategyServer,
			expectedOutput:  nil,
		},
		{
			name:            "ResolveStrategyClient returns client",
			resolveStrategy: models.ResolveStrategyClient,
			expectedOutput:  clientReq,
		},
		{
			name:            "ResolveStrategyInteractive default server",
			resolveStrategy: models.ResolveStrategyInteractive,
			input:           "\n", // Enter pressed, default server
			expectedOutput:  nil,
		},
		{
			name:            "ResolveStrategyInteractive choose client",
			resolveStrategy: models.ResolveStrategyInteractive,
			input:           "client\n",
			expectedOutput:  clientReq,
		},
		{
			name:            "ResolveStrategyInteractive invalid then client",
			resolveStrategy: models.ResolveStrategyInteractive,
			input:           "invalid\nclient\n",
			expectedOutput:  clientReq,
		},
		{
			name:            "Unknown resolve strategy returns error",
			resolveStrategy: "unknown",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))

			got, err := client.ResolveConflictUsernamePassword(ctx, reader, server, clientReq, tt.resolveStrategy)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, got)
		})
	}
}
