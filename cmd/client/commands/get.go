package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
)

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
			ctx := cmd.Context()

			if secretName == "" {
				return fmt.Errorf("--secret-name is required")
			}

			if serverURL == "" {
				// Local mode
				if clientPubKeyFile == "" {
					return fmt.Errorf("--pubkey is required in local mode")
				}

				sqlite, err := db.New("sqlite", "client.db")
				if err != nil {
					return fmt.Errorf("failed to connect to DB: %w", err)
				}
				defer sqlite.Close()

				reader := repositories.NewEncryptedSecretReadRepository(sqlite)

				crypt, err := cryptor.New(
					cryptor.WithPrivateKeyFromFile(clientPubKeyFile),
				)
				if err != nil {
					return fmt.Errorf("failed to init decryptor: %w", err)
				}

				secret, err := reader.Get(ctx, secretName)
				if err != nil {
					return fmt.Errorf("failed to read secret: %w", err)
				}
				if secret == nil {
					fmt.Println("Secret not found")
					return nil
				}

				enc := &cryptor.Encrypted{
					Ciphertext: secret.Ciphertext,
					AESKeyEnc:  secret.AESKeyEnc,
				}
				plaintext, err := crypt.Decrypt(enc)
				if err != nil {
					return fmt.Errorf("failed to decrypt secret: %w", err)
				}

				fmt.Println(string(plaintext))
				return nil
			}

			// Remote mode
			if clientPubKeyFile == "" {
				return fmt.Errorf("--pubkey is required in remote mode")
			}

			crypt, err := cryptor.New(
				cryptor.WithPublicKeyFromCert(clientPubKeyFile),
			)
			if err != nil {
				return fmt.Errorf("failed to init cryptor: %w", err)
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

				facade := facades.NewSecretHTTPReadFacade(httpClient)

				secret, err := facade.Get(ctx, secretName)
				if err != nil {
					return fmt.Errorf("failed to get secret via HTTP: %w", err)
				}
				if secret == nil {
					fmt.Println("Secret not found")
					return nil
				}

				enc := &cryptor.Encrypted{
					Ciphertext: secret.Ciphertext,
					AESKeyEnc:  secret.AESKeyEnc,
				}
				plaintext, err := crypt.Decrypt(enc)
				if err != nil {
					return fmt.Errorf("failed to decrypt secret: %w", err)
				}

				fmt.Println(string(plaintext))
				return nil

			case scheme.GRPC:
				grpcConn, err := grpc.New(serverURL,
					grpc.WithAuthToken(authToken),
					grpc.WithRetryPolicy(grpc.RetryPolicy{
						Count:   3,
						Wait:    500 * time.Millisecond,
						MaxWait: 5 * time.Second,
					}),
				)
				if err != nil {
					return fmt.Errorf("failed to create gRPC client: %w", err)
				}
				defer grpcConn.Close()

				facade := facades.NewSecretGRPCReadFacade(grpcConn)

				secret, err := facade.Get(ctx, secretName)
				if err != nil {
					return fmt.Errorf("failed to get secret via gRPC: %w", err)
				}
				if secret == nil {
					fmt.Println("Secret not found")
					return nil
				}

				enc := &cryptor.Encrypted{
					Ciphertext: secret.Ciphertext,
					AESKeyEnc:  secret.AESKeyEnc,
				}
				plaintext, err := crypt.Decrypt(enc)
				if err != nil {
					return fmt.Errorf("failed to decrypt secret: %w", err)
				}

				fmt.Println(string(plaintext))
				return nil

			default:
				return fmt.Errorf("unsupported scheme: %s", schemeType)
			}
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret to retrieve (required)")
	cmd.Flags().StringVar(&serverURL, "server", "", "Server URL (e.g., http://localhost:8080 or grpc://localhost:50051)")
	cmd.Flags().StringVar(&clientPubKeyFile, "pubkey", "", "Path to public/private key file (required)")
	cmd.Flags().StringVar(&authToken, "token", "", "Bearer token for remote server (optional)")

	return cmd
}
