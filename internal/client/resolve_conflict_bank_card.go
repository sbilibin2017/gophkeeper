package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func ResolveConflictBankCard(
	ctx context.Context,
	reader *bufio.Reader,
	server *models.BankCardResponse,
	client *models.BankCardAddRequest,
	resolveStrategy string,
) (*models.BankCardAddRequest, error) {
	fmt.Println("Starting sync using strategy:", resolveStrategy)

	switch resolveStrategy {
	case models.ResolveStrategyServer:
		return nil, nil
	case models.ResolveStrategyClient:
		return client, nil
	case models.ResolveStrategyInteractive:
		fmt.Println("Conflict detected between server and client versions.")

		serverJSON, err := json.MarshalIndent(server, "  ", "  ")
		if err != nil {
			fmt.Println("Error formatting server JSON:", err)
		} else {
			fmt.Println("Server version:")
			fmt.Println(string(serverJSON))
		}

		clientJSON, err := json.MarshalIndent(client, "  ", "  ")
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
