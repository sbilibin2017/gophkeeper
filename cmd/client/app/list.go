package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newClientListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List saved secrets with optional filtering and sorting",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server_url")
			hmacKey, _ := cmd.Flags().GetString("hmac_key")
			rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa_public_key")
			secretType, _ := cmd.Flags().GetString("type")
			interactive, _ := cmd.Flags().GetBool("interactive")

			if interactive {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter secret type to filter (leave empty for all): ")
				input, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				secretType = strings.TrimSpace(input)
			}

			cmd.Printf("Listing secrets from server: %s\n", serverURL)
			if secretType != "" {
				cmd.Printf("Filter by type: %s\n", secretType)
			} else {
				cmd.Println("No type filter applied, listing all secrets")
			}
			cmd.Println("HMAC key:", hmacKey)
			cmd.Println("RSA public key path:", rsaPublicKeyPath)

			return nil
		},
	}

	cmd.Flags().String("server_url", "", "Server URL")
	cmd.Flags().String("hmac_key", "", "HMAC key")
	cmd.Flags().String("rsa_public_key", "", "Path to RSA public key")

	cmd.Flags().Bool("interactive", false, "Enable interactive input for type filter")

	cmd.Flags().String("type", "", "Filter secrets by type")

	_ = cmd.MarkFlagRequired("server_url")

	return cmd
}
