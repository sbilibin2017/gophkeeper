package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

// RegisterAddBinaryCommand registers the 'add-binary-secret' command.
func RegisterAddBinaryCommand(root *cobra.Command) {
	var (
		secretName    string
		dataPath      string
		meta          string
		authURL       string
		tlsClientCert string
		tlsClientKey  string
	)

	cmd := &cobra.Command{
		Use:   "add-binary-secret",
		Short: "Add a binary secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			var opts []configs.ClientConfigOpt

			opts = append(opts, configs.WithClientConfigDB())

			schm := scheme.GetSchemeFromURL(authURL)

			switch schm {
			case scheme.HTTP, scheme.HTTPS:
				httpOpts := []clients.HTTPClientOption{}
				if tlsClientCert != "" && tlsClientKey != "" {
					httpOpts = append(httpOpts, clients.WithHTTPTLSClientCert(tlsClientCert, tlsClientKey))
				}
				opts = append(opts, configs.WithClientConfigHTTPClient(authURL, httpOpts...))

			case scheme.GRPC:
				grpcOpts := []clients.GRPCClientOption{}
				if tlsClientCert != "" && tlsClientKey != "" {
					grpcOpts = append(grpcOpts, clients.WithGRPCTLSClientCert(tlsClientCert, tlsClientKey))
				}
				opts = append(opts, configs.WithClientConfigGRPCClient(authURL, grpcOpts...))

			default:
				return errors.New("unsupported scheme: " + schm)
			}

			cfg, err := configs.NewClientConfig(opts...)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

			data, err := os.ReadFile(dataPath)
			if err != nil {
				return err
			}

			req := models.BinaryAddRequest{
				SecretName: secretName,
				Data:       data,
			}
			if meta != "" {
				req.Meta = &meta
			}

			ctx := cmd.Context()

			if cfg.DB != nil {
				err := client.AddBinaryLocal(ctx, cfg.DB, req)
				if err != nil {
					return err
				}
				cmd.Println("Binary secret added locally")
				return nil
			}

			return errors.New("no local DB configured for adding binary secret")
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&dataPath, "data-path", "", "Path to binary data file")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("data-path")

	root.AddCommand(cmd)
}
