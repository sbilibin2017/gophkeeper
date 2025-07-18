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

func RegisterDeleteSecretCommand(root *cobra.Command) {
	var (
		secretType    string
		secretName    string
		authURL       string
		tlsClientCert string
		tlsClientKey  string
		token         string
	)

	cmd := &cobra.Command{
		Use:   "delete-secret",
		Short: "Delete a secret by type and name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validation.ValidateSecretType(secretType); err != nil {
				return err
			}
			if err := validation.ValidateSecretName(secretName); err != nil {
				return err
			}

			cfg, err := config.NewClientConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return err
			}

			mode := scheme.GetSchemeFromURL(authURL)

			switch secretType {
			case models.SecretTypeBankCard:
				switch mode {
				case scheme.HTTP, scheme.HTTPS:
					err = client.DeleteBankCardHTTP(cmd.Context(), cfg.HTTPClient, token, secretName)
				case scheme.GRPC:
					clientGRPC := pb.NewBankCardDeleteServiceClient(cfg.GRPCClient)
					err = client.DeleteBankCardGRPC(cmd.Context(), clientGRPC, token, secretName)
				default:
					err = client.DeleteBankCardLocal(cmd.Context(), cfg.DB, secretName)
				}

			case models.SecretTypeBinary:
				switch mode {
				case scheme.HTTP, scheme.HTTPS:
					err = client.DeleteBinaryHTTP(cmd.Context(), cfg.HTTPClient, token, secretName)
				case scheme.GRPC:
					clientGRPC := pb.NewBinaryDeleteServiceClient(cfg.GRPCClient)
					err = client.DeleteBinaryGRPC(cmd.Context(), clientGRPC, token, secretName)
				default:
					err = client.DeleteBinaryLocal(cmd.Context(), cfg.DB, secretName)
				}

			case models.SecretTypeText:
				switch mode {
				case scheme.HTTP, scheme.HTTPS:
					err = client.DeleteTextHTTP(cmd.Context(), cfg.HTTPClient, token, secretName)
				case scheme.GRPC:
					clientGRPC := pb.NewTextDeleteServiceClient(cfg.GRPCClient)
					err = client.DeleteTextGRPC(cmd.Context(), clientGRPC, token, secretName)
				default:
					err = client.DeleteTextLocal(cmd.Context(), cfg.DB, secretName)
				}

			case models.SecretTypeUsernamePassword:
				switch mode {
				case scheme.HTTP, scheme.HTTPS:
					err = client.DeleteUsernamePasswordHTTP(cmd.Context(), cfg.HTTPClient, token, secretName)
				case scheme.GRPC:
					clientGRPC := pb.NewUsernamePasswordDeleteServiceClient(cfg.GRPCClient)
					err = client.DeleteUsernamePasswordGRPC(cmd.Context(), clientGRPC, token, secretName)
				default:
					err = client.DeleteUsernamePasswordLocal(cmd.Context(), cfg.DB, secretName)
				}

			default:
				return fmt.Errorf("unsupported secret type: %s", secretType)
			}

			if err != nil {
				return err
			}

			cmd.Printf("Secret deleted successfully: type=%s name=%s\n", secretType, secretName)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&secretType, "secret-type", "", "Type of secret (bankcard, binary, text, usernamepassword)")
	flags.StringVar(&secretName, "secret-name", "", "Name of the secret to delete")
	flags.StringVar(&authURL, "auth-url", "", "Service URL (e.g. http://, https://, grpc://) to detect transport")
	flags.StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")
	flags.StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS private key file (optional)")
	flags.StringVar(&token, "token", "", "Authentication token")

	root.AddCommand(cmd)
}
