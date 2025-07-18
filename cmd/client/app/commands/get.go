package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/validation"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func RegisterGetSecretCommand(root *cobra.Command) {
	var (
		secretType    string
		secretName    string
		authURL       string
		tlsClientCert string
		tlsClientKey  string
		token         string
	)

	cmd := &cobra.Command{
		Use:   "get-secret",
		Short: "Get a secret by type and name",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := validation.ValidateSecretType(secretType); err != nil {
				return fmt.Errorf("invalid secret type: %w", err)
			}
			if err := validation.ValidateSecretName(secretName); err != nil {
				return fmt.Errorf("invalid secret name: %w", err)
			}

			cfg, err := config.NewClientConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return err
			}

			mode := scheme.GetSchemeFromURL(authURL)

			var res any

			switch secretType {
			case models.SecretTypeBankCard:
				switch mode {
				case scheme.HTTP, scheme.HTTPS:
					res, err = client.GetBankCardHTTP(ctx, cfg.HTTPClient, token, secretName)
				case scheme.GRPC:
					clientGRPC := pb.NewBankCardGetServiceClient(cfg.GRPCClient)
					res, err = client.GetBankCardGRPC(ctx, clientGRPC, token, secretName)
				default:
					res, err = client.GetBankCardLocal(ctx, cfg.DB, secretName)
				}

			case models.SecretTypeBinary:
				switch mode {
				case scheme.HTTP, scheme.HTTPS:
					res, err = client.GetBinaryHTTP(ctx, cfg.HTTPClient, token, secretName)
				case scheme.GRPC:
					clientGRPC := pb.NewBinaryGetServiceClient(cfg.GRPCClient)
					res, err = client.GetBinaryGRPC(ctx, clientGRPC, token, secretName)
				default:
					res, err = client.GetBinaryLocal(ctx, cfg.DB, secretName)
				}

			case models.SecretTypeText:
				switch mode {
				case scheme.HTTP, scheme.HTTPS:
					res, err = client.GetTextHTTP(ctx, cfg.HTTPClient, token, secretName)
				case scheme.GRPC:
					clientGRPC := pb.NewTextGetServiceClient(cfg.GRPCClient)
					res, err = client.GetTextGRPC(ctx, clientGRPC, token, secretName)
				default:
					res, err = client.GetTextLocal(ctx, cfg.DB, secretName)
				}

			case models.SecretTypeUsernamePassword:
				switch mode {
				case scheme.HTTP, scheme.HTTPS:
					res, err = client.GetUsernamePasswordHTTP(ctx, cfg.HTTPClient, token, secretName)
				case scheme.GRPC:
					clientGRPC := pb.NewUsernamePasswordGetServiceClient(cfg.GRPCClient)
					res, err = client.GetUsernamePasswordGRPC(ctx, clientGRPC, token, secretName)
				default:
					res, err = client.GetUsernamePasswordLocal(ctx, cfg.DB, secretName)
				}

			default:
				return fmt.Errorf("unsupported secret-type %q", secretType)
			}

			if err != nil {
				return err
			}

			cmd.Println(res)
			return nil
		},
	}

	cmd.Flags().StringVar(&secretType, "secret-type", "", fmt.Sprintf("Type of secret (%s, %s, %s, %s)",
		models.SecretTypeBankCard, models.SecretTypeBinary, models.SecretTypeText, models.SecretTypeUsernamePassword))
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Service URL (e.g. http://, https://, grpc://) to detect transport")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS private key file (optional)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	_ = cmd.MarkFlagRequired("secret-type")
	_ = cmd.MarkFlagRequired("secret-name")

	root.AddCommand(cmd)
}
