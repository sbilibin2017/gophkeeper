package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newClientGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get secret data from the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server_url")
			hmacKey, _ := cmd.Flags().GetString("hmac_key")
			rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa_public_key")
			secretID, _ := cmd.Flags().GetString("secret_id")
			interactive, _ := cmd.Flags().GetBool("interactive")

			if interactive && secretID == "" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter secret ID (leave empty to fetch all): ")
				input, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				secretID = strings.TrimSpace(input)
			}

			if serverURL == "" {
				return fmt.Errorf("--server_url is required")
			}

			cmd.Println("HMAC key:", hmacKey)
			cmd.Println("RSA public key path:", rsaPublicKeyPath)

			return nil
		},
	}

	cmd.Flags().String("server_url", "", "Server URL")
	cmd.Flags().String("hmac_key", "", "HMAC key")
	cmd.Flags().String("rsa_public_key", "", "Path to RSA public key")
	cmd.Flags().String("secret_id", "", "Optional secret ID to fetch specific secret")
	cmd.Flags().Bool("interactive", false, "Enable interactive input for secret ID")

	_ = cmd.MarkFlagRequired("server_url")

	return cmd
}
