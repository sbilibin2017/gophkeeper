package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/validation"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func RegisterListSecretsCommand(root *cobra.Command) {
	var (
		secretType    string
		authURL       string
		tlsClientCert string
		tlsClientKey  string
		token         string
	)

	cmd := &cobra.Command{
		Use:   "list-secrets",
		Short: "List names of secrets of a specified type",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if secretType != "" {
				if err := validation.ValidateSecretType(secretType); err != nil {
					return err
				}
			}

			cfg, err := config.NewClientConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return err
			}

			mode := scheme.GetSchemeFromURL(authURL)

			var names []string

			appendNames := func(items interface{}) {
				switch list := items.(type) {
				case []*models.BankCardResponse:
					for _, item := range list {
						names = append(names, item.SecretName)
					}
				case []*models.BinaryResponse:
					for _, item := range list {
						names = append(names, item.SecretName)
					}
				case []*models.TextResponse:
					for _, item := range list {
						names = append(names, item.SecretName)
					}
				case []*models.UsernamePasswordResponse:
					for _, item := range list {
						names = append(names, item.SecretName)
					}
				}
			}

			for _, t := range []string{
				models.SecretTypeBankCard,
				models.SecretTypeBinary,
				models.SecretTypeText,
				models.SecretTypeUsernamePassword,
			} {
				if secretType != "" && secretType != t {
					continue
				}

				switch t {
				case models.SecretTypeBankCard:
					switch mode {
					case scheme.HTTP, scheme.HTTPS:
						items, err := client.ListBankCardsHTTP(ctx, cfg.HTTPClient, token)
						if err != nil {
							return fmt.Errorf("failed to list bank cards: %v", err)
						}
						appendNames(items)
					case scheme.GRPC:
						clientGRPC := pb.NewBankCardListServiceClient(cfg.GRPCClient)
						items, err := client.ListBankCardsGRPC(ctx, clientGRPC, token)
						if err != nil {
							return fmt.Errorf("failed to list bank cards: %v", err)
						}
						appendNames(items)
					default:
						items, err := client.ListBankCardsLocal(ctx, cfg.DB)
						if err != nil {
							return fmt.Errorf("failed to list local bank cards: %v", err)
						}
						appendNames(items)
					}

				case models.SecretTypeBinary:
					switch mode {
					case scheme.HTTP, scheme.HTTPS:
						items, err := client.ListBinaryHTTP(ctx, cfg.HTTPClient, token)
						if err != nil {
							return fmt.Errorf("failed to list binary secrets: %v", err)
						}
						appendNames(items)
					case scheme.GRPC:
						clientGRPC := pb.NewBinaryListServiceClient(cfg.GRPCClient)
						items, err := client.ListBinaryGRPC(ctx, clientGRPC, token)
						if err != nil {
							return fmt.Errorf("failed to list binary secrets: %v", err)
						}
						appendNames(items)
					default:
						items, err := client.ListBinaryLocal(ctx, cfg.DB)
						if err != nil {
							return fmt.Errorf("failed to list local binary secrets: %v", err)
						}
						appendNames(items)
					}

				case models.SecretTypeText:
					switch mode {
					case scheme.HTTP, scheme.HTTPS:
						items, err := client.ListTextHTTP(ctx, cfg.HTTPClient, token)
						if err != nil {
							return fmt.Errorf("failed to list text secrets: %v", err)
						}
						appendNames(items)
					case scheme.GRPC:
						clientGRPC := pb.NewTextListServiceClient(cfg.GRPCClient)
						items, err := client.ListTextGRPC(ctx, clientGRPC, token)
						if err != nil {
							return fmt.Errorf("failed to list text secrets: %v", err)
						}
						appendNames(items)
					default:
						items, err := client.ListTextLocal(ctx, cfg.DB)
						if err != nil {
							return fmt.Errorf("failed to list local text secrets: %v", err)
						}
						appendNames(items)
					}

				case models.SecretTypeUsernamePassword:
					switch mode {
					case scheme.HTTP, scheme.HTTPS:
						items, err := client.ListUsernamePasswordHTTP(ctx, cfg.HTTPClient, token)
						if err != nil {
							return fmt.Errorf("failed to list username/password secrets: %v", err)
						}
						appendNames(items)
					case scheme.GRPC:
						clientGRPC := pb.NewUsernamePasswordListServiceClient(cfg.GRPCClient)
						items, err := client.ListUsernamePasswordGRPC(ctx, clientGRPC, token)
						if err != nil {
							return fmt.Errorf("failed to list username/password secrets: %v", err)
						}
						appendNames(items)
					default:
						items, err := client.ListUsernamePasswordLocal(ctx, cfg.DB)
						if err != nil {
							return fmt.Errorf("failed to list local username/password secrets: %v", err)
						}
						appendNames(items)
					}

				default:
					return fmt.Errorf("unsupported secret type %q", t)
				}
			}

			for _, name := range names {
				cmd.Println(name)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&secretType, "secret-type", "", fmt.Sprintf("Type of secrets to list (%s, %s, %s, %s)", models.SecretTypeBankCard, models.SecretTypeBinary, models.SecretTypeText, models.SecretTypeUsernamePassword))
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Service URL (e.g. http://, https://, grpc://) to detect transport")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS private key file (optional)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	root.AddCommand(cmd)
}
