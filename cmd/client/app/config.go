package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	var token string
	var serverURL string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure client parameters",
		Example: `  gophkeeper config --token mytoken123
  gophkeeper config --server-url https://example.com
  gophkeeper config --token mytoken123 --server-url https://example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" && serverURL == "" {
				return fmt.Errorf("you must provide at least one flag: --token or --server-url")
			}

			if token != "" {
				err := os.Setenv("GOPHKEEPER_TOKEN", token)
				if err != nil {
					return fmt.Errorf("failed to set environment variable GOPHKEEPER_TOKEN: %w", err)
				}
				fmt.Println("Token saved to GOPHKEEPER_TOKEN")
			}

			if serverURL != "" {
				err := os.Setenv("GOPHKEEPER_SERVER_URL", serverURL)
				if err != nil {
					return fmt.Errorf("failed to set environment variable GOPHKEEPER_SERVER_URL: %w", err)
				}
				fmt.Println("Server URL saved to GOPHKEEPER_SERVER_URL")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL")

	return cmd
}
