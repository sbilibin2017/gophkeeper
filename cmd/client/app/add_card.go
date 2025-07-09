package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newAddCardCommand() *cobra.Command {
	var number, expiry, cvv, token, serverURL string
	var interactive bool
	var metas []string

	cmd := &cobra.Command{
		Use:   "add-card",
		Short: "Add bank card details with optional metadata",
		Example: `  gophkeeper add-card --number 4111111111111111 --expiry 12/25 --cvv 123 --meta owner=John --token mytoken --server-url https://example.com
  gophkeeper add-card --interactive
  gophkeeper add-card --number 5555444433332222 --expiry 01/26 --cvv 999 --meta category=business --server-url https://example.com --token mytoken`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter card number: ")
				inputNumber, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				number = strings.TrimSpace(inputNumber)

				fmt.Print("Enter expiration date: ")
				inputExpiry, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				expiry = strings.TrimSpace(inputExpiry)

				fmt.Print("Enter CVV code: ")
				inputCVV, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				cvv = strings.TrimSpace(inputCVV)

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

			if number == "" || expiry == "" || cvv == "" {
				return fmt.Errorf("parameters number, expiry, and cvv are required")
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

			fmt.Printf("Card added: %s (CVV: %s, expiry: %s), token: %s, server: %s\nMetadata: %+v\n",
				number, cvv, expiry, token, serverURL, metadata)

			// TODO: implement card storage including metadata

			return nil
		},
	}

	cmd.Flags().StringVar(&number, "number", "", "Card number")
	cmd.Flags().StringVar(&expiry, "expiry", "", "Expiration date")
	cmd.Flags().StringVar(&cvv, "cvv", "", "CVV code")
	cmd.Flags().StringSliceVar(&metas, "meta", []string{}, "Metadata key=value pairs (can be specified multiple times)")
	cmd.Flags().StringVar(&token, "token", "", "Authorization token (can be set via GOPHKEEPER_TOKEN env variable)")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL (can be set via GOPHKEEPER_SERVER_URL env variable)")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive input mode")

	return cmd
}
