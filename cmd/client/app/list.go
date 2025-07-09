package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	var secretType, token, serverURL string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display saved data",
		Example: `  gophkeeper list
  gophkeeper list --type login
  gophkeeper list --token mytoken --server-url https://example.com
  gophkeeper list --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter authorization token (leave empty to use GOPHKEEPER_TOKEN environment variable): ")
				inputToken, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				token = strings.TrimSpace(inputToken)

				fmt.Print("Enter server URL (leave empty to use GOPHKEEPER_SERVER_URL environment variable): ")
				inputServerURL, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				serverURL = strings.TrimSpace(inputServerURL)
			}

			if token == "" {
				token = os.Getenv("GOPHKEEPER_TOKEN")
			}
			if serverURL == "" {
				serverURL = os.Getenv("GOPHKEEPER_SERVER_URL")
			}

			if token == "" || serverURL == "" {
				return fmt.Errorf("you must provide both token and server URL via flags, interactively, or environment variables")
			}

			if secretType != "" {
				fmt.Printf("Displaying only data of type: %s\n", secretType)
				// TODO: filter and display data of the given type using token and serverURL
			} else {
				fmt.Println("Displaying all saved data")
				// TODO: display all data using token and serverURL
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&secretType, "type", "", "Filter by secret type (login, text, binary, card)")
	cmd.Flags().StringVar(&token, "token", "", "Authorization token (can also be provided via GOPHKEEPER_TOKEN environment variable)")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL (can also be provided via GOPHKEEPER_SERVER_URL environment variable)")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Enable interactive input mode")

	return cmd
}
