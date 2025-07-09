package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newAddLoginPasswordCommand() *cobra.Command {
	var login, password, token, serverURL string
	var interactive bool
	var metas []string

	cmd := &cobra.Command{
		Use:   "add-login-password",
		Short: "Add login and password with optional metadata",
		Example: `  gophkeeper add-login-password --login user123 --password secret --meta site=example.com --token mytoken --server-url https://example.com
  gophkeeper add-login-password --interactive
  gophkeeper add-login-password --login backupuser --password pass123 --meta category=work --server-url https://example.com --token mytoken`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter login: ")
				inputLogin, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				login = strings.TrimSpace(inputLogin)

				fmt.Print("Enter password: ")
				inputPassword, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				password = strings.TrimSpace(inputPassword)

				fmt.Println("Enter metadata key=value pairs one by one. Leave empty to finish:")
				for {
					fmt.Print("> ")
					line, err := reader.ReadString('\n')
					if err != nil {
						return err
					}
					line = strings.TrimSpace(line)
					if line == "" {
						break
					}
					metas = append(metas, line)
				}

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

			if login == "" || password == "" {
				return fmt.Errorf("parameters login and password are required")
			}
			if token == "" || serverURL == "" {
				return fmt.Errorf("token and server URL must be provided via flags, interactive input, or environment variables")
			}

			metadata := map[string]string{}
			for _, m := range metas {
				parts := strings.SplitN(m, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid metadata format, expected key=value but got: %s", m)
				}
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				metadata[key] = value
			}

			fmt.Printf("Login added: %s, password: %s, token: %s, server: %s\nMetadata: %+v\n",
				login, strings.Repeat("*", len(password)), token, serverURL, metadata)

			return nil
		},
	}

	cmd.Flags().StringVar(&login, "login", "", "Username")
	cmd.Flags().StringVar(&password, "password", "", "Password")
	cmd.Flags().StringSliceVar(&metas, "meta", []string{}, "Metadata key=value pairs (can be specified multiple times)")
	cmd.Flags().StringVar(&token, "token", "", "Authorization token (can be set via GOPHKEEPER_TOKEN env variable)")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL (can be set via GOPHKEEPER_SERVER_URL env variable)")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive input mode")

	return cmd
}
