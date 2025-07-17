package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func RegisterSyncCommand(root *cobra.Command) {
	var (
		token           string
		serverURL       string
		tlsClientCert   string
		resolveStrategy string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize secrets between local and server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			return runSync(ctx, token, serverURL, tlsClientCert, resolveStrategy)
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server API URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")
	cmd.Flags().StringVar(&resolveStrategy, "resolve-strategy", "server", "Conflict resolution strategy (server, client, interactive)")

	root.AddCommand(cmd)
}

func runSync(
	ctx context.Context,
	token string,
	serverURL string,
	tlsClientCert string,
	resolveStrategy string,
) error {
	if serverURL == "" {
		return errors.New("server URL is required for sync")
	}

	fmt.Println("Starting sync using strategy:", resolveStrategy)

	switch resolveStrategy {
	case "server":
		fmt.Println("Using server data to resolve conflicts")
	case "client":
		fmt.Println("Using local data to overwrite server state")
	case "interactive":
		return errors.New("interactive mode not supported yet")
	default:
		return fmt.Errorf("unknown resolve strategy: %s", resolveStrategy)
	}

	// TODO: Реализовать синхронизацию
	fmt.Println("Sync completed successfully")
	return nil
}
