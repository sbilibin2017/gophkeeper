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

func TestResolveConflictText(t *testing.T) {
	ctx := context.Background()

	server := &models.TextResponse{
		SecretName: "serverText",
		Content:    "server content",
		Meta:       nil,
	}

	clientReq := &models.TextAddRequest{
		SecretName: "clientText",
		Content:    "client content",
		Meta:       nil,
	}

	tests := []struct {
		name            string
		resolveStrategy string
		input           string
		expectedOutput  *models.TextAddRequest
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
			input:           "\n", // default to server
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
			input:           "wrong\nclient\n",
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

			got, err := client.ResolveConflictText(ctx, reader, server, clientReq, tt.resolveStrategy)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, got)
		})
	}
}
