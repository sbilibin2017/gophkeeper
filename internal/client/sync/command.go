package sync

import (
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/address"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

// NewCommand returns the "sync" CLI command.
func NewCommand() *cobra.Command {
	var (
		serverURL string
		token     string
		mode      string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize secrets with the server",
		Long: `Synchronize secrets between the client and the Gophkeeper backend.
Requires an authentication token.

Supported sync modes:
- client: automatic client-side conflict resolution
- interactive: manual conflict resolution with user interaction

Examples of usage and supported server URL schemes:
- http:// for HTTP
- grpc:// for gRPC
`,
		Example: `  # Sync secrets using HTTP with client mode (default)
  gophkeeper sync --token <token>

  # Sync secrets using gRPC with interactive mode
  gophkeeper sync --token <token> --server-url grpc://localhost:50051 --mode interactive
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			addr := address.New(serverURL)
			if addr.Address == "" || addr.Scheme == "" {
				return errors.New("invalid server URL format")
			}

			// Default to client mode if not set
			if mode == "" {
				mode = models.SyncModeClient
			}

			switch addr.Scheme {
			case address.SchemeHTTP, address.SchemeHTTPS:
				err := RunHTTP(ctx, token, addr.Address, mode)
				if err != nil {
					return fmt.Errorf("failed to sync secrets over HTTP: %w", err)
				}
			case address.SchemeGRPC:
				err := RunGRPC(ctx, token, addr.Address, mode)
				if err != nil {
					return fmt.Errorf("failed to sync secrets over gRPC: %w", err)
				}
			default:
				return address.ErrUnsupportedScheme
			}

			cmd.Println("Synchronization completed successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "http://localhost:8080", "Server address (scheme://host:port)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&mode, "mode", models.SyncModeClient, "Sync mode (client or interactive)")

	return cmd
}
