package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/facades"

	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/spf13/cobra"
)

// NewGetCommand returns a cobra command that retrieves a secret by name.
//
// The command supports both local and remote secret retrieval. If no --server flag is provided,
// the secret is fetched from the local SQLite database using the private key for decryption.
// If the --server flag is provided, the secret is fetched from the server (HTTP or gRPC),
// using the provided public key file for cryptographic operations.
//
// Supported flags:
// - --secret-name / -n: name of the secret to retrieve (required)
// - --server: server URL to fetch secret from (optional; if empty, local mode is used)
// - --pubkey: path to the client public or private key file (required for both modes)
// - --token: authorization token for server requests (optional; used only in remote mode)
func NewGetCommand() *cobra.Command {
	var (
		secretName       string
		serverURL        string
		clientPubKeyFile string
		authToken        string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Gets secret",
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
				secret, err := secretReader.Get(ctx, secretName)
				if err != nil {
					return fmt.Errorf("failed to get secret: %w", err)
				}
				if secret == nil {
					fmt.Println("Secret not found")
					return nil
				}
				fmt.Println(*secret)
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

				secret, err := secretReader.Get(ctx, secretName)
				if err != nil {
					return fmt.Errorf("failed to get secret via HTTP: %w", err)
				}
				if secret == nil {
					fmt.Println("Secret not found")
					return nil
				}
				fmt.Println(*secret)
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

				secret, err := secretReader.Get(ctx, secretName)
				if err != nil {
					return fmt.Errorf("failed to get secret via gRPC: %w", err)
				}
				if secret == nil {
					fmt.Println("Secret not found")
					return nil
				}
				fmt.Println(*secret)
				return nil

			default:
				return fmt.Errorf("unsupported or missing scheme in server URL")
			}
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret to retrieve")
	cmd.Flags().StringVar(&serverURL, "server", "", "Server URL (http://, https://, grpc://)")
	cmd.Flags().StringVar(&clientPubKeyFile, "pubkey", "", "Client public key file path")
	cmd.Flags().StringVar(&authToken, "token", "", "Authorization token for server requests (optional)")

	return cmd
}
