package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/resolver"
	"github.com/spf13/cobra"
)

// NewSyncCommand creates a Cobra command for syncing secrets between the client and server.
//
// The sync operation can run in multiple modes:
//   - "client": Synchronizes secrets from the client to the server.
//   - "server": Currently a no-op (reserved for future server-side sync implementation).
//   - "interactive": Provides an interactive mode for resolving sync conflicts.
//
// The command supports communication with the server over HTTP(S) or gRPC,
// determined by the scheme in the server URL. It uses a local SQLite database
// ("client.db") for storing client secrets during synchronization.
//
// Flags:
//
//	--server: URL of the server including scheme (e.g., http://, https://, grpc://).
//	--pubkey: Path to the client's public key file used for encryption/decryption.
//	--token:  Optional authorization token for server requests.
//	--mode:   Sync mode to operate in; one of "client", "server", or "interactive" (default "client").
//
// Returns an error if initialization fails or if the sync process encounters issues.
func NewSyncCommand() *cobra.Command {
	var (
		serverURL        string
		clientPubKeyFile string
		authToken        string
		mode             string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync secrets between client and server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			switch mode {
			case models.SyncModeServer:
				return nil

			case models.SyncModeClient, models.SyncModeInteractive:
				schm := scheme.GetSchemeFromURL(serverURL)

				switch schm {
				case scheme.HTTP:
					dbConn, err := db.New("sqlite", "client.db")
					if err != nil {
						return fmt.Errorf("failed to open local DB: %w", err)
					}
					defer dbConn.Close()

					clientReader := repositories.NewEncryptedSecretReadRepository(dbConn)

					httpClient, err := http.New(serverURL,
						http.WithAuthToken(authToken),
						http.WithRetryPolicy(http.RetryPolicy{
							Count:   3,
							Wait:    500 * time.Millisecond,
							MaxWait: 5 * time.Second,
						}),
					)
					if err != nil {
						return fmt.Errorf("failed to create HTTP client: %w", err)
					}

					serverWriter := facades.NewSecretHTTPWriteFacade(httpClient)
					serverReader := facades.NewSecretHTTPReadFacade(httpClient)

					cryptor, err := cryptor.New(
						cryptor.WithPublicKeyFromCert(clientPubKeyFile),
					)
					if err != nil {
						return fmt.Errorf("failed to init cryptor: %w", err)
					}

					clientSecretReader := client.NewSecretReader(clientReader, cryptor)
					serverSecretReader := client.NewSecretReader(serverReader, cryptor)
					serverSecretWriter := client.NewSecretWriter(serverWriter, cryptor)

					r := resolver.NewResolver(clientSecretReader, serverSecretReader, serverSecretWriter)

					switch mode {
					case models.SyncModeClient:
						if err := r.ResolveClient(ctx); err != nil {
							return fmt.Errorf("client sync failed: %w", err)
						}
					case models.SyncModeInteractive:
						if err := r.ResolveInteractive(ctx, cmd.InOrStdin()); err != nil {
							return fmt.Errorf("interactive sync failed: %w", err)
						}
					}

					err = repositories.DropEncryptedSecretsTable(cmd.Context(), dbConn)
					if err != nil {
						return err
					}

				case scheme.GRPC:
					dbConn, err := db.New("sqlite", "client.db")
					if err != nil {
						return fmt.Errorf("failed to open local DB: %w", err)
					}
					defer dbConn.Close()

					clientReader := repositories.NewEncryptedSecretReadRepository(dbConn)

					grpcClient, err := grpc.New(serverURL,
						grpc.WithAuthToken(authToken),
						grpc.WithRetryPolicy(grpc.RetryPolicy{
							Count:   3,
							Wait:    500 * time.Millisecond,
							MaxWait: 5 * time.Second,
						}),
					)
					if err != nil {
						return fmt.Errorf("failed to create HTTP client: %w", err)
					}

					serverWriter := facades.NewSecretGRPCWriteFacade(grpcClient)
					serverReader := facades.NewSecretGRPCReadFacade(grpcClient)

					r := resolver.NewResolver(clientReader, serverReader, serverWriter)

					switch mode {
					case models.SyncModeClient:
						if err := r.ResolveClient(ctx); err != nil {
							return fmt.Errorf("client sync failed: %w", err)
						}
					case models.SyncModeInteractive:
						if err := r.ResolveInteractive(ctx, cmd.InOrStdin()); err != nil {
							return fmt.Errorf("interactive sync failed: %w", err)
						}
					}

					err = repositories.DropEncryptedSecretsTable(cmd.Context(), dbConn)
					if err != nil {
						return err
					}

				default:
					return fmt.Errorf("unsupported scheme: %s", schm)
				}

			default:
				return fmt.Errorf("invalid sync mode: %s", mode)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "Server URL (http://, https://, grpc://)")
	cmd.Flags().StringVar(&clientPubKeyFile, "pubkey", "", "Client public key file path")
	cmd.Flags().StringVar(&authToken, "token", "", "Authorization token for server requests (optional)")
	cmd.Flags().StringVar(&mode, "mode", models.SyncModeClient, "Sync mode: client, server, or interactive")

	return cmd
}
