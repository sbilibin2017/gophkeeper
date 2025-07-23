package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Long:  "Register a new user by providing credentials and TLS client authentication details",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		serverURL, _ := cmd.Flags().GetString("server-url")
		certFile, _ := cmd.Flags().GetString("cert-file")
		certKey, _ := cmd.Flags().GetString("cert-key")

		schemeType := scheme.GetSchemeFromURL(serverURL)

		switch schemeType {
		case scheme.HTTP, scheme.HTTPS:
			token, err := runRegisterHTTP(ctx, serverURL, certFile, certKey, username, password)
			if err != nil {
				return err
			}
			cmd.Println(token)

		case scheme.GRPC:
			token, err := runRegisterGRPC(ctx, serverURL, certFile, certKey, username, password)
			if err != nil {
				return err
			}
			cmd.Println(token)

		default:
			return fmt.Errorf("unsupported scheme: %s", schemeType)
		}

		return nil
	},
}

func runRegisterHTTP(
	ctx context.Context,
	serverURL, certFile, certKey, username, password string,
) (string, error) {
	client, err := http.New(
		serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    500 * time.Millisecond,
			MaxWait: 2 * time.Second,
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP client: %w", err)
	}

	authFacade := facades.NewAuthHTTPFacade(client)
	req := &models.AuthRequest{Login: username, Password: password}
	resp, err := authFacade.Register(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

func runRegisterGRPC(
	ctx context.Context,
	serverURL, certFile, certKey, username, password string,
) (string, error) {
	conn, err := grpc.New(
		serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    500 * time.Millisecond,
			MaxWait: 2 * time.Second,
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	authFacade := facades.NewAuthGRPCFacade(conn)
	req := &models.AuthRequest{Login: username, Password: password}
	resp, err := authFacade.Register(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

func init() {
	registerCmd.Flags().StringP("username", "u", "", "Username (required)")
	registerCmd.Flags().StringP("password", "p", "", "Password (required)")
	registerCmd.Flags().String("server-url", "", "Server URL (required)")
	registerCmd.Flags().String("cert-file", "", "Path to client certificate file (required)")
	registerCmd.Flags().String("cert-key", "", "Path to client private key file (required)")

	_ = registerCmd.MarkFlagRequired("username")
	_ = registerCmd.MarkFlagRequired("password")
	_ = registerCmd.MarkFlagRequired("server-url")
	_ = registerCmd.MarkFlagRequired("cert-file")
	_ = registerCmd.MarkFlagRequired("cert-key")
}
