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
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/spf13/cobra"
)

// NewListCommand returns a cobra command that lists all secrets.
//
// The command supports both local and remote secret listing. If no --server flag is provided,
// the secrets are fetched from the local SQLite database using the private key for decryption.
// If the --server flag is provided, the secrets are fetched from the server (HTTP or gRPC),
// using the provided public key file for cryptographic operations.
//
// Supported flags:
// - --server: server URL to fetch secrets from (optional; if empty, local mode is used)
// - --pubkey: path to the client public or private key file (required for both modes)
// - --token: authorization token for server requests (optional; used only in remote mode)
func NewListCommand() *cobra.Command {
	var (
		serverURL        string
		clientPubKeyFile string
		authToken        string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if serverURL == "" || clientPubKeyFile == "" {
				db, err := db.New("sqlite", "client.db")
				if err != nil {
					return fmt.Errorf("failed to connect to DB: %w", err)
				}
				defer db.Close()

				reader := repositories.NewEncryptedSecretReadRepository(db)

				cryptor, err := cryptor.New(
					cryptor.WithPrivateKeyFromFile(clientPubKeyFile),
				)
				if err != nil {
					return fmt.Errorf("failed to init local decryptor: %w", err)
				}

				secretReader := client.NewSecretReader(reader, cryptor)

				secrets, err := secretReader.List(ctx)
				if err != nil {
					return fmt.Errorf("failed to list secrets: %w", err)
				}
				if len(secrets) == 0 {
					fmt.Println("No secrets found")
					return nil
				}
				for _, s := range secrets {
					fmt.Println(s)
				}
				return nil
			}

			schemeType := scheme.GetSchemeFromURL(serverURL)
			switch schemeType {
			case scheme.HTTP, scheme.HTTPS:
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

				cryptor, err := cryptor.New(
					cryptor.WithPublicKeyFromCert(clientPubKeyFile),
				)
				if err != nil {
					return fmt.Errorf("failed to init cryptor: %w", err)
				}

				httpReader := facades.NewSecretHTTPReadFacade(httpClient)

				secretReader := client.NewSecretReader(httpReader, cryptor)

				secrets, err := secretReader.List(ctx)
				if err != nil {
					return fmt.Errorf("failed to list secrets via HTTP: %w", err)
				}
				if len(secrets) == 0 {
					fmt.Println("No secrets found")
					return nil
				}
				for _, s := range secrets {
					fmt.Println(s)
				}
				return nil

			case scheme.GRPC:
				grpcClient, err := grpc.New(serverURL,
					grpc.WithAuthToken(authToken),
					grpc.WithRetryPolicy(grpc.RetryPolicy{
						Count:   3,
						Wait:    500 * time.Millisecond,
						MaxWait: 5 * time.Second,
					}),
				)
				if err != nil {
					return fmt.Errorf("failed to create gRPC connection: %w", err)
				}
				defer grpcClient.Close()

				cryptor, err := cryptor.New(
					cryptor.WithPublicKeyFromCert(clientPubKeyFile),
				)
				if err != nil {
					return fmt.Errorf("failed to init cryptor: %w", err)
				}

				grpcReader := facades.NewSecretGRPCReadFacade(grpcClient)

				secretReader := client.NewSecretReader(grpcReader, cryptor)

				secrets, err := secretReader.List(ctx)
				if err != nil {
					return fmt.Errorf("failed to list secrets via gRPC: %w", err)
				}
				if len(secrets) == 0 {
					fmt.Println("No secrets found")
					return nil
				}
				for _, s := range secrets {
					fmt.Println(s)
				}
				return nil

			default:
				return fmt.Errorf("unsupported or missing scheme in server URL")
			}
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "Server URL (http://, https://, grpc://)")
	cmd.Flags().StringVar(&clientPubKeyFile, "pubkey", "", "Client public key file path")
	cmd.Flags().StringVar(&authToken, "token", "", "Authorization token for server requests (optional)")

	return cmd
}
