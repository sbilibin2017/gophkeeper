package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newSyncCommand() *cobra.Command {
	var token, serverURL, resolver string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize data with the server",
		Example: `  gophkeeper sync --token mytoken --server-url https://example.com --resolver server
  gophkeeper sync --interactive
  gophkeeper sync --resolver interactive --token mytoken --server-url https://example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter authorization token (leave empty to use GOPHKEEPER_TOKEN environment variable): ")
				inputToken, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				inputToken = strings.TrimSpace(inputToken)
				if inputToken != "" {
					token = inputToken
				}

				fmt.Print("Enter server URL (leave empty to use GOPHKEEPER_SERVER_URL environment variable): ")
				inputServerURL, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				inputServerURL = strings.TrimSpace(inputServerURL)
				if inputServerURL != "" {
					serverURL = inputServerURL
				}

				fmt.Print("Enter conflict resolution strategy (server/client/interactive): ")
				inputResolver, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				inputResolver = strings.TrimSpace(inputResolver)
				if inputResolver != "" {
					resolver = inputResolver
				}
			}

			if token == "" {
				token = os.Getenv("GOPHKEEPER_TOKEN")
			}
			if serverURL == "" {
				serverURL = os.Getenv("GOPHKEEPER_SERVER_URL")
			}

			if token == "" || serverURL == "" {
				return fmt.Errorf("you must provide both a token and a server URL via flags, interactive mode, or environment variables")
			}

			validResolvers := map[string]bool{
				"server":      true,
				"client":      true,
				"interactive": true,
				"":            true, // allow empty (default behavior)
			}
			if !validResolvers[resolver] {
				return fmt.Errorf("invalid --resolver value: %s. Allowed values: server, client, interactive", resolver)
			}

			fmt.Printf("Synchronizing with server %s using token %s\n", serverURL, token)
			fmt.Printf("Conflict resolution strategy: %s\n", resolver)

			// TODO: Implement your synchronization and conflict resolution logic here

			if resolver == "interactive" {
				fmt.Println("Interactive conflict resolution activated")
				// Implement interactive conflict handling here
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Authorization token (can be provided via GOPHKEEPER_TOKEN environment variable)")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL (can be provided via GOPHKEEPER_SERVER_URL environment variable)")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Enable interactive input mode")
	cmd.Flags().StringVar(&resolver, "resolver", "", "Conflict resolution strategy (server, client, interactive)")

	return cmd
}
