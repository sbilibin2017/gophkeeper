package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// User authentication
func newLoginCommand() *cobra.Command {
	var username, password, serverURL string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate user",
		Example: `  gophkeeper login --username alice --password secret123 --server-url https://example.com
  gophkeeper login --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter username: ")
				userInput, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				username = strings.TrimSpace(userInput)

				fmt.Print("Enter password: ")
				passInput, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				password = strings.TrimSpace(passInput)

				fmt.Print("Enter server URL: ")
				urlInput, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				serverURL = strings.TrimSpace(urlInput)
			}

			if username == "" || password == "" {
				return fmt.Errorf("username and password cannot be empty")
			}

			if serverURL == "" {
				return fmt.Errorf("server-url must be provided")
			}

			// TODO: Implement authentication logic with username, password, and serverURL
			fmt.Printf("Authenticating user: %s with password: %s on server: %s\n",
				username, strings.Repeat("*", len(password)), serverURL)

			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username")
	cmd.Flags().StringVar(&password, "password", "", "Password")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Enable interactive input mode")

	return cmd
}
