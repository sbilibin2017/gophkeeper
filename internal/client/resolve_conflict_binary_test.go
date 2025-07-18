package client_test

import (
	"bufio"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func TestResolveConflictBinary(t *testing.T) {
	ctx := context.Background()

	now := time.Now()

	server := &models.BinaryResponse{
		SecretName:  "serverSecret",
		SecretOwner: "serverOwner",
		Data:        []byte{0x01, 0x02, 0x03},
		Meta:        nil,
		UpdatedAt:   now,
	}

	clientReq := &models.BinaryAddRequest{
		SecretName: "clientSecret",
		Data:       []byte{0x0A, 0x0B, 0x0C},
		Meta:       nil,
	}

	tests := []struct {
		name            string
		resolveStrategy string
		input           string // input for interactive mode
		expectedOutput  *models.BinaryAddRequest
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
			input:           "\n", // default server
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
			name:            "Unknown strategy returns error",
			resolveStrategy: "unknown",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))

			got, err := client.ResolveConflictBinary(ctx, reader, server, clientReq, tt.resolveStrategy)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, got)
		})
	}
}
