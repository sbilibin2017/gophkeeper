package commands

import (
	"bufio"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/models"

	"github.com/sbilibin2017/gophkeeper/internal/validation"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func RegisterSyncCommand(root *cobra.Command) {
	var (
		serverURL       string
		tlsClientCert   string
		tlsClientKey    string
		token           string
		resolveStrategy string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize secrets between local and server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := validation.ValidateResolveStrategy(resolveStrategy); err != nil {
				return fmt.Errorf("invalid resolve strategy: %w", err)
			}

			cfg, err := config.NewClientConfig(serverURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return err
			}

			// Detect scheme from serverURL
			protocol := scheme.GetSchemeFromURL(serverURL)

			var httpClient *resty.Client
			var grpcConn *grpc.ClientConn

			switch protocol {
			case scheme.GRPC:
				grpcConn = cfg.GRPCClient
				defer grpcConn.Close()
			case scheme.HTTP, scheme.HTTPS:
				httpClient = cfg.HTTPClient
			default:
				return fmt.Errorf("unsupported protocol scheme: %s", protocol)
			}

			reader := bufio.NewReader(cmd.InOrStdin())

			// --- BankCards ---
			localBankCards, err := client.ListBankCardsLocal(ctx, cfg.DB)
			if err != nil {
				return fmt.Errorf("failed to list local bank cards: %w", err)
			}

			for _, localCard := range localBankCards {
				var serverCard *models.BankCardResponse

				if protocol == scheme.GRPC {
					bankCardGetClient := pb.NewBankCardGetServiceClient(grpcConn)
					serverCard, err = client.GetBankCardGRPC(ctx, bankCardGetClient, token, localCard.SecretName)
				} else {
					serverCard, err = client.GetBankCardHTTP(ctx, httpClient, token, localCard.SecretName)
				}
				if err != nil {
					return fmt.Errorf("failed to get bank card '%s' from server: %w", localCard.SecretName, err)
				}

				resolvedCard, err := client.ResolveConflictBankCard(ctx, reader, serverCard, localCard, resolveStrategy)
				if err != nil {
					return fmt.Errorf("conflict resolution failed for bank card '%s': %w", localCard.SecretName, err)
				}

				if resolvedCard == nil {
					continue
				}

				if protocol == scheme.GRPC {
					bankCardAddClient := pb.NewBankCardAddServiceClient(grpcConn)
					err = client.AddBankCardGRPC(ctx, bankCardAddClient, token, *resolvedCard)
				} else {
					err = client.AddBankCardHTTP(ctx, httpClient, token, *resolvedCard)
				}
				if err != nil {
					return fmt.Errorf("failed to update bank card '%s' on server: %w", resolvedCard.SecretName, err)
				}
			}

			// --- Binary ---
			localBinaries, err := client.ListBinaryLocal(ctx, cfg.DB)
			if err != nil {
				return fmt.Errorf("failed to list local binary secrets: %w", err)
			}

			for _, localBinary := range localBinaries {
				var serverBinary *models.BinaryResponse

				if protocol == scheme.GRPC {
					binaryGetClient := pb.NewBinaryGetServiceClient(grpcConn)
					serverBinary, err = client.GetBinaryGRPC(ctx, binaryGetClient, token, localBinary.SecretName)
				} else {
					serverBinary, err = client.GetBinaryHTTP(ctx, httpClient, token, localBinary.SecretName)
				}
				if err != nil {
					return fmt.Errorf("failed to get binary secret '%s' from server: %w", localBinary.SecretName, err)
				}

				resolvedBinary, err := client.ResolveConflictBinary(ctx, reader, serverBinary, localBinary, resolveStrategy)
				if err != nil {
					return fmt.Errorf("conflict resolution failed for binary secret '%s': %w", localBinary.SecretName, err)
				}

				if resolvedBinary == nil {
					continue
				}

				if protocol == scheme.GRPC {
					binaryAddClient := pb.NewBinaryAddServiceClient(grpcConn)
					err = client.AddBinaryGRPC(ctx, binaryAddClient, token, *resolvedBinary)
				} else {
					err = client.AddBinaryHTTP(ctx, httpClient, token, *resolvedBinary)
				}
				if err != nil {
					return fmt.Errorf("failed to update binary secret '%s' on server: %w", resolvedBinary.SecretName, err)
				}
			}

			// --- Text ---
			localTexts, err := client.ListTextLocal(ctx, cfg.DB)
			if err != nil {
				return fmt.Errorf("failed to list local text secrets: %w", err)
			}

			for _, localText := range localTexts {
				var serverText *models.TextResponse

				if protocol == scheme.GRPC {
					textGetClient := pb.NewTextGetServiceClient(grpcConn)
					serverText, err = client.GetTextGRPC(ctx, textGetClient, token, localText.SecretName)
				} else {
					serverText, err = client.GetTextHTTP(ctx, httpClient, token, localText.SecretName)
				}
				if err != nil {
					return fmt.Errorf("failed to get text secret '%s' from server: %w", localText.SecretName, err)
				}

				resolvedText, err := client.ResolveConflictText(ctx, reader, serverText, localText, resolveStrategy)
				if err != nil {
					return fmt.Errorf("conflict resolution failed for text secret '%s': %w", localText.SecretName, err)
				}

				if resolvedText == nil {
					continue
				}

				if protocol == scheme.GRPC {
					textAddClient := pb.NewTextAddServiceClient(grpcConn)
					err = client.AddTextGRPC(ctx, textAddClient, token, *resolvedText)
				} else {
					err = client.AddTextHTTP(ctx, httpClient, token, *resolvedText)
				}
				if err != nil {
					return fmt.Errorf("failed to update text secret '%s' on server: %w", resolvedText.SecretName, err)
				}
			}

			// --- UsernamePassword ---
			localUserPasses, err := client.ListUsernamePasswordLocal(ctx, cfg.DB)
			if err != nil {
				return fmt.Errorf("failed to list local username-password secrets: %w", err)
			}

			for _, localUserPass := range localUserPasses {
				var serverUserPass *models.UsernamePasswordResponse

				if protocol == scheme.GRPC {
					upGetClient := pb.NewUsernamePasswordGetServiceClient(grpcConn)
					serverUserPass, err = client.GetUsernamePasswordGRPC(ctx, upGetClient, token, localUserPass.SecretName)
				} else {
					serverUserPass, err = client.GetUsernamePasswordHTTP(ctx, httpClient, token, localUserPass.SecretName)
				}
				if err != nil {
					return fmt.Errorf("failed to get username-password secret '%s' from server: %w", localUserPass.SecretName, err)
				}

				resolvedUserPass, err := client.ResolveConflictUsernamePassword(ctx, reader, serverUserPass, localUserPass, resolveStrategy)
				if err != nil {
					return fmt.Errorf("conflict resolution failed for username-password secret '%s': %w", localUserPass.SecretName, err)
				}

				if resolvedUserPass == nil {
					continue
				}

				if protocol == scheme.GRPC {
					upAddClient := pb.NewUsernamePasswordAddServiceClient(grpcConn)
					err = client.AddUsernamePasswordGRPC(ctx, upAddClient, token, *resolvedUserPass)
				} else {
					err = client.AddUsernamePasswordHTTP(ctx, httpClient, token, *resolvedUserPass)
				}
				if err != nil {
					return fmt.Errorf("failed to update username-password secret '%s' on server: %w", resolvedUserPass.SecretName, err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server API URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file (optional)")
	cmd.Flags().StringVar(&resolveStrategy, "resolve-strategy", "server", "Conflict resolution strategy (server, client, interactive)")

	root.AddCommand(cmd)
}
