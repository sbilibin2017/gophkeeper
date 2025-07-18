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

func TestResolveConflictBankCard(t *testing.T) {
	ctx := context.Background()

	server := &models.BankCardResponse{
		SecretName: "serverName",
		Number:     "1111222233334444",
		Owner:      "Server Owner",
		Exp:        "12/34",
		CVV:        "123",
	}

	clientReq := &models.BankCardAddRequest{
		SecretName: "clientName",
		Number:     "5555666677778888",
		Owner:      "Client Owner",
		Exp:        "11/22",
		CVV:        "456",
	}

	tests := []struct {
		name            string
		resolveStrategy string
		input           string // for interactive input
		expectedOutput  *models.BankCardAddRequest
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
			name:            "ResolveStrategyInteractive chooses server by default",
			resolveStrategy: models.ResolveStrategyInteractive,
			input:           "\n", // default = server
			expectedOutput:  nil,
		},
		{
			name:            "ResolveStrategyInteractive chooses client",
			resolveStrategy: models.ResolveStrategyInteractive,
			input:           "client\n",
			expectedOutput:  clientReq,
		},
		{
			name:            "ResolveStrategyInteractive invalid input then client",
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

			got, err := client.ResolveConflictBankCard(ctx, reader, server, clientReq, tt.resolveStrategy)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, got)
		})
	}
}
