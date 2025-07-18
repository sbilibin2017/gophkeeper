package commands

import (
	"bufio"

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
				return err
			}

			cfg, err := config.NewClientConfig(serverURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return err
			}

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
				return nil // undefined protocol
			}

			reader := bufio.NewReader(cmd.InOrStdin())

			// --- Sync Bank Cards ---
			localBankCards, err := client.ListBankCardsLocal(ctx, cfg.DB)
			if err != nil {
				return err
			}

			for _, localCard := range localBankCards {
				var serverCard *models.BankCardResponse

				if protocol == scheme.GRPC {
					serverCard, err = client.GetBankCardGRPC(ctx, pb.NewBankCardGetServiceClient(grpcConn), token, localCard.SecretName)
				} else {
					serverCard, err = client.GetBankCardHTTP(ctx, httpClient, token, localCard.SecretName)
				}
				if err != nil {
					return err
				}

				resolvedCard, err := client.ResolveConflictBankCard(ctx, reader, serverCard, localCard, resolveStrategy)
				if err != nil {
					return err
				}

				if resolvedCard == nil {
					continue
				}

				if protocol == scheme.GRPC {
					err = client.AddBankCardGRPC(ctx, pb.NewBankCardAddServiceClient(grpcConn), token, *resolvedCard)
				} else {
					err = client.AddBankCardHTTP(ctx, httpClient, token, *resolvedCard)
				}
				if err != nil {
					return err
				}
			}

			if err := client.DropBankCardRequestTable(ctx, cfg.DB); err != nil {
				return err
			}

			// --- Sync Binary ---
			localBinaries, err := client.ListBinaryLocal(ctx, cfg.DB)
			if err != nil {
				return err
			}

			for _, localBin := range localBinaries {
				var serverBin *models.BinaryResponse

				if protocol == scheme.GRPC {
					serverBin, err = client.GetBinaryGRPC(ctx, pb.NewBinaryGetServiceClient(grpcConn), token, localBin.SecretName)
				} else {
					serverBin, err = client.GetBinaryHTTP(ctx, httpClient, token, localBin.SecretName)
				}
				if err != nil {
					return err
				}

				resolvedBin, err := client.ResolveConflictBinary(ctx, reader, serverBin, localBin, resolveStrategy)
				if err != nil {
					return err
				}

				if resolvedBin == nil {
					continue
				}

				if protocol == scheme.GRPC {
					err = client.AddBinaryGRPC(ctx, pb.NewBinaryAddServiceClient(grpcConn), token, *resolvedBin)
				} else {
					err = client.AddBinaryHTTP(ctx, httpClient, token, *resolvedBin)
				}
				if err != nil {
					return err
				}
			}

			if err := client.DropBinaryRequestTable(ctx, cfg.DB); err != nil {
				return err
			}

			// --- Sync Text ---
			localTexts, err := client.ListTextLocal(ctx, cfg.DB)
			if err != nil {
				return err
			}

			for _, localText := range localTexts {
				var serverText *models.TextResponse

				if protocol == scheme.GRPC {
					serverText, err = client.GetTextGRPC(ctx, pb.NewTextGetServiceClient(grpcConn), token, localText.SecretName)
				} else {
					serverText, err = client.GetTextHTTP(ctx, httpClient, token, localText.SecretName)
				}
				if err != nil {
					return err
				}

				resolvedText, err := client.ResolveConflictText(ctx, reader, serverText, localText, resolveStrategy)
				if err != nil {
					return err
				}

				if resolvedText == nil {
					continue
				}

				if protocol == scheme.GRPC {
					err = client.AddTextGRPC(ctx, pb.NewTextAddServiceClient(grpcConn), token, *resolvedText)
				} else {
					err = client.AddTextHTTP(ctx, httpClient, token, *resolvedText)
				}
				if err != nil {
					return err
				}
			}

			if err := client.DropTextRequestTable(ctx, cfg.DB); err != nil {
				return err
			}

			// --- Sync Username-Password ---
			localUPs, err := client.ListUsernamePasswordLocal(ctx, cfg.DB)
			if err != nil {
				return err
			}

			for _, localUP := range localUPs {
				var serverUP *models.UsernamePasswordResponse

				if protocol == scheme.GRPC {
					serverUP, err = client.GetUsernamePasswordGRPC(ctx, pb.NewUsernamePasswordGetServiceClient(grpcConn), token, localUP.SecretName)
				} else {
					serverUP, err = client.GetUsernamePasswordHTTP(ctx, httpClient, token, localUP.SecretName)
				}
				if err != nil {
					return err
				}

				resolvedUP, err := client.ResolveConflictUsernamePassword(ctx, reader, serverUP, localUP, resolveStrategy)
				if err != nil {
					return err
				}

				if resolvedUP == nil {
					continue
				}

				if protocol == scheme.GRPC {
					err = client.AddUsernamePasswordGRPC(ctx, pb.NewUsernamePasswordAddServiceClient(grpcConn), token, *resolvedUP)
				} else {
					err = client.AddUsernamePasswordHTTP(ctx, httpClient, token, *resolvedUP)
				}
				if err != nil {
					return err
				}
			}

			if err := client.DropUsernamePasswordRequestTable(ctx, cfg.DB); err != nil {
				return err
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
