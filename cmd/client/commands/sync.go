package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/resolver"
	"github.com/spf13/cobra"
)

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
