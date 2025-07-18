package client

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func ResolveConflictBinary(
	ctx context.Context,
	reader *bufio.Reader,
	server *models.BinaryResponse,
	client *models.BinaryAddRequest,
	resolveStrategy string,
) (*models.BinaryAddRequest, error) {
	fmt.Println("Starting sync using strategy:", resolveStrategy)

	switch resolveStrategy {
	case models.ResolveStrategyServer:
		return nil, nil
	case models.ResolveStrategyClient:
		return client, nil
	case models.ResolveStrategyInteractive:
		fmt.Println("Conflict detected between server and client versions.")

		// Print Server JSON with base64 data string for readability
		serverCopy := struct {
			SecretName  string  `json:"secret_name"`
			SecretOwner string  `json:"secret_owner"`
			Data        string  `json:"data"` // base64 encoded
			Meta        *string `json:"meta,omitempty"`
			UpdatedAt   string  `json:"updated_at"`
		}{
			SecretName:  server.SecretName,
			SecretOwner: server.SecretOwner,
			Data:        base64.StdEncoding.EncodeToString(server.Data),
			Meta:        server.Meta,
			UpdatedAt:   server.UpdatedAt.String(),
		}
		serverJSON, err := json.MarshalIndent(serverCopy, "  ", "  ")
		if err != nil {
			fmt.Println("Error formatting server JSON:", err)
		} else {
			fmt.Println("Server version:")
			fmt.Println(string(serverJSON))
		}

		// Print Client JSON with base64 encoded Data
		clientCopy := struct {
			SecretName string  `json:"secret_name"`
			Data       string  `json:"data"` // base64 encoded
			Meta       *string `json:"meta,omitempty"`
		}{
			SecretName: client.SecretName,
			Data:       base64.StdEncoding.EncodeToString(client.Data),
			Meta:       client.Meta,
		}
		clientJSON, err := json.MarshalIndent(clientCopy, "  ", "  ")
		if err != nil {
			fmt.Println("Error formatting client JSON:", err)
		} else {
			fmt.Println("Client version:")
			fmt.Println(string(clientJSON))
		}

		for {
			fmt.Print("Choose which version to keep ([server]/client): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			input = strings.TrimSpace(strings.ToLower(input))

			if input == models.ResolveStrategyServer || input == "" {
				return nil, nil
			}
			if input == models.ResolveStrategyClient {
				return client, nil
			}

			fmt.Println("Invalid input, please enter 'server' or 'client'.")
		}
	default:
		return nil, fmt.Errorf("unknown resolve strategy: %s", resolveStrategy)
	}
}
