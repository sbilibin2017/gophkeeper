package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newAddTextCommand() *cobra.Command {
	var data, token, serverURL string
	var interactive bool
	var metas []string

	cmd := &cobra.Command{
		Use:   "add-text",
		Short: "Add arbitrary text data",
		Example: `  gophkeeper add-text --data "some secret text" --meta note=personal --token mytoken --server-url https://example.com
  gophkeeper add-text --interactive
  gophkeeper add-text --data "backup notes" --meta category=work --server-url https://example.com --token mytoken`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter text data: ")
				inputData, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				data = strings.TrimSpace(inputData)

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

			if data == "" {
				return fmt.Errorf("parameter data is required")
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

			fmt.Printf("Text added: %s\nToken: %s\nServer: %s\nMetadata: %+v\n",
				data, token, serverURL, metadata)

			// TODO: сохранить текст и метаданные на сервере

			return nil
		},
	}

	cmd.Flags().StringVar(&data, "data", "", "Text data")
	cmd.Flags().StringSliceVar(&metas, "meta", []string{}, "Metadata key=value pairs (can be specified multiple times)")
	cmd.Flags().StringVar(&token, "token", "", "Authorization token (can be set via GOPHKEEPER_TOKEN env variable)")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL (can be set via GOPHKEEPER_SERVER_URL env variable)")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive input mode")

	return cmd
}
