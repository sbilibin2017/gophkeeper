package app

import (
	"context"
	"errors"
	"os"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/sbilibin2017/gophkeeper/pkg/grpc" // grpc client package (adjust import if needed)
	"github.com/spf13/cobra"
)

// newAddCommand creates a cobra command for adding new data/secrets
func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add --server-url <url> --type <type> --file <path> --interactive --hmac-key <key> --rsa-public-key <path>",
		Short: "Add new secret to the client from a file or interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			serverURL, _ := cmd.Flags().GetString("server-url")
			stype, _ := cmd.Flags().GetString("type")
			filePath, _ := cmd.Flags().GetString("file")
			interactive, _ := cmd.Flags().GetBool("interactive")
			hmacKey, _ := cmd.Flags().GetString("hmac-key")
			rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa-public-key")

			if filePath == "" && !interactive {
				return errors.New("either --file or --interactive must be specified")
			}
			if filePath != "" && interactive {
				return errors.New("--file and --interactive cannot be used together")
			}

			config, err := configs.NewClientConfig(
				configs.WithClient(serverURL),
				configs.WithHMACEncoder(hmacKey),
				configs.WithRSAEncoder(rsaPublicKeyPath),
			)
			if err != nil {
				return err
			}

			switch stype {
			case models.TypeLoginPassword:
				if config.HTTPClient != nil {
					if filePath != "" {
						file, err := os.Open(filePath)
						if err != nil {
							return err
						}
						defer file.Close()
						return services.AddLoginPasswordHTTPWithFile(ctx, config.HTTPClient, file, config.HMACEncoder, config.RSAEncoder)
					}
					return services.AddLoginPasswordHTTPWithInteractive(ctx, config.HTTPClient, config.HMACEncoder, config.RSAEncoder)
				} else if config.GRPCClient != nil {
					grpcClient := grpc.NewAddServiceClient(config.GRPCClient)
					if filePath != "" {
						file, err := os.Open(filePath)
						if err != nil {
							return err
						}
						defer file.Close()
						return services.AddLoginPasswordGRPCFile(ctx, grpcClient, file, config.HMACEncoder, config.RSAEncoder)
					}
					return services.AddLoginPasswordGRPCInteractive(ctx, grpcClient, config.HMACEncoder, config.RSAEncoder)
				}
				return errors.New("unsupported client type")

			case models.TypeText:
				if config.HTTPClient != nil {
					if filePath != "" {
						file, err := os.Open(filePath)
						if err != nil {
							return err
						}
						defer file.Close()
						return services.AddTextHTTPFile(ctx, config.HTTPClient, file, config.HMACEncoder, config.RSAEncoder)
					}
					return services.AddTextHTTPInteractive(ctx, config.HTTPClient, config.HMACEncoder, config.RSAEncoder)
				} else if config.GRPCClient != nil {
					grpcClient := grpc.NewAddServiceClient(config.GRPCClient)
					if filePath != "" {
						file, err := os.Open(filePath)
						if err != nil {
							return err
						}
						defer file.Close()
						return services.AddTextGRPCFile(ctx, grpcClient, file, config.HMACEncoder, config.RSAEncoder)
					}
					return services.AddTextGRPCInteractive(ctx, grpcClient, config.HMACEncoder, config.RSAEncoder)
				}
				return errors.New("unsupported client type")

			case models.TypeBinary:
				if config.HTTPClient != nil {
					if filePath != "" {
						file, err := os.Open(filePath)
						if err != nil {
							return err
						}
						defer file.Close()
						return services.AddBinaryHTTPFile(ctx, config.HTTPClient, file, config.HMACEncoder, config.RSAEncoder)
					}
					return errors.New("interactive mode not supported for binary type")
				} else if config.GRPCClient != nil {
					grpcClient := grpc.NewAddServiceClient(config.GRPCClient)
					if filePath != "" {
						file, err := os.Open(filePath)
						if err != nil {
							return err
						}
						defer file.Close()
						return services.AddBinaryGRPCFile(ctx, grpcClient, file, config.HMACEncoder, config.RSAEncoder)
					}
					return errors.New("interactive mode not supported for binary type")
				}
				return errors.New("unsupported client type")

			case models.TypeCard:
				if config.HTTPClient != nil {
					if filePath != "" {
						file, err := os.Open(filePath)
						if err != nil {
							return err
						}
						defer file.Close()
						return services.AddCardHTTPFile(ctx, config.HTTPClient, file, config.HMACEncoder, config.RSAEncoder)
					}
					return services.AddCardHTTPInteractive(ctx, config.HTTPClient, config.HMACEncoder, config.RSAEncoder)
				} else if config.GRPCClient != nil {
					grpcClient := grpc.NewAddServiceClient(config.GRPCClient)
					if filePath != "" {
						file, err := os.Open(filePath)
						if err != nil {
							return err
						}
						defer file.Close()
						return services.AddCardGRPCFile(ctx, grpcClient, file, config.HMACEncoder, config.RSAEncoder)
					}
					return services.AddCardGRPCInteractive(ctx, grpcClient, config.HMACEncoder, config.RSAEncoder)
				}
				return errors.New("unsupported client type")

			default:
				return errors.New("unsupported secret type")
			}
		},
	}

	cmd.Flags().StringP("server-url", "s", "", "Server URL")
	cmd.Flags().StringP("type", "t", "", "Secret type (login_password, text, binary, card)")
	cmd.Flags().StringP("file", "f", "", "Input file path")
	cmd.Flags().BoolP("interactive", "i", false, "Enable interactive input mode")
	cmd.Flags().String("hmac-key", "", "HMAC encryption key")
	cmd.Flags().String("rsa-public-key", "", "Path to RSA public key")

	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}
