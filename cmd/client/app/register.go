package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newRegisterCommand() *cobra.Command {
	var username, password, serverURL string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		Example: `  gophkeeper register --username alice --password secret123 --server-url https://example.com
  gophkeeper register --interactive`,
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

			// TODO: Implement registration logic using username, password, and serverURL
			fmt.Printf("Registering user: %s with password: %s at server: %s\n",
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
