package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newAddBinaryCommand() *cobra.Command {
	var filePath, token, serverURL string
	var interactive bool
	var metas []string

	cmd := &cobra.Command{
		Use:   "add-binary",
		Short: "Add binary data from file with optional text metadata",
		Example: `  gophkeeper add-binary --file ./path/to/file.bin --meta site=example.com --meta user=john --token mytoken --server-url https://example.com
  gophkeeper add-binary --interactive
  gophkeeper add-binary --file backup.bin --meta codes="1234,5678,9012" --server-url https://example.com --token mytoken`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter path to binary file: ")
				inputFile, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				filePath = strings.TrimSpace(inputFile)

				fmt.Print("Enter metadata as key=value pairs (comma separated, e.g. site=example.com,user=john), or leave empty: ")
				inputMeta, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				inputMeta = strings.TrimSpace(inputMeta)
				if inputMeta != "" {
					for _, pair := range strings.Split(inputMeta, ",") {
						metas = append(metas, strings.TrimSpace(pair))
					}
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

			if filePath == "" {
				return fmt.Errorf("parameter file is required")
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

			fmt.Printf("Binary file added: %s\nToken: %s\nServer: %s\nMetadata: %+v\n",
				filePath, token, serverURL, metadata)

			// TODO: read the file, attach metadata, and send to the server

			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to binary file")
	cmd.Flags().StringSliceVar(&metas, "meta", []string{}, "Metadata key=value pairs (can be specified multiple times)")
	cmd.Flags().StringVar(&token, "token", "", "Authorization token (can be set via GOPHKEEPER_TOKEN env variable)")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL (can be set via GOPHKEEPER_SERVER_URL env variable)")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive input mode")

	return cmd
}
