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
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
)

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
			ctx := cmd.Context()

			var (
				secrets []*models.EncryptedSecret
				crypt   *cryptor.Cryptor
				err     error
			)

			if serverURL == "" {
				// Local mode
				dbConn, err := db.New("sqlite", "client.db")
				if err != nil {
					return fmt.Errorf("failed to connect to DB: %w", err)
				}
				defer dbConn.Close()

				reader := repositories.NewEncryptedSecretReadRepository(dbConn)

				crypt, err = cryptor.New(
					cryptor.WithPrivateKeyFromFile(clientPubKeyFile),
				)
				if err != nil {
					return fmt.Errorf("failed to init decryptor: %w", err)
				}

				secrets, err = reader.List(ctx)
				if err != nil {
					return fmt.Errorf("failed to list secrets locally: %w", err)
				}
			} else {
				// Remote mode
				crypt, err = cryptor.New(
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
					secrets, err = facade.List(ctx)
					if err != nil {
						return fmt.Errorf("failed to list secrets via HTTP: %w", err)
					}

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
						return fmt.Errorf("failed to create gRPC connection: %w", err)
					}
					defer grpcConn.Close()

					facade := facades.NewSecretGRPCReadFacade(grpcConn)
					secrets, err = facade.List(ctx)
					if err != nil {
						return fmt.Errorf("failed to list secrets via gRPC: %w", err)
					}

				default:
					return fmt.Errorf("unsupported or missing scheme in server URL")
				}
			}

			if len(secrets) == 0 {
				fmt.Println("No secrets found")
				return nil
			}

			for _, s := range secrets {
				enc := &cryptor.Encrypted{
					Ciphertext: s.Ciphertext,
					AESKeyEnc:  s.AESKeyEnc,
				}
				plaintext, err := crypt.Decrypt(enc)
				if err != nil {
					continue
				}
				cmd.Println(plaintext)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "Server URL (http://, https://, grpc://)")
	cmd.Flags().StringVar(&clientPubKeyFile, "pubkey", "", "Client public/private key file path (required)")
	cmd.Flags().StringVar(&authToken, "token", "", "Authorization token for server requests (optional)")

	return cmd
}
