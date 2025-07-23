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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login as an existing user",
	Long:  "Authenticate an existing user using credentials and TLS certificate",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		serverURL, _ := cmd.Flags().GetString("server-url")
		certFile, _ := cmd.Flags().GetString("cert-file")

		schemeType := scheme.GetSchemeFromURL(serverURL)

		switch schemeType {
		case scheme.HTTP, scheme.HTTPS:
			token, err := runLoginHTTP(ctx, serverURL, certFile, username, password)
			if err != nil {
				return err
			}
			cmd.Println(token)
		case scheme.GRPC:
			token, err := runLoginGRPC(ctx, serverURL, certFile, username, password)
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

func runLoginHTTP(
	ctx context.Context,
	serverURL, certFile, username, password string,
) (string, error) {
	client, err := http.New(
		serverURL,
		http.WithTLSCert(certFile),
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
	resp, err := authFacade.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

func runLoginGRPC(
	ctx context.Context,
	serverURL, certFile, username, password string,
) (string, error) {
	conn, err := grpc.New(
		serverURL,
		grpc.WithTLSCert(certFile),
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
	resp, err := authFacade.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

func init() {
	loginCmd.Flags().StringP("username", "u", "", "Username (required)")
	loginCmd.Flags().StringP("password", "p", "", "Password (required)")
	loginCmd.Flags().String("server-url", "", "Server URL (required)")
	loginCmd.Flags().String("cert-file", "", "Path to client certificate file (required)")

	_ = loginCmd.MarkFlagRequired("username")
	_ = loginCmd.MarkFlagRequired("password")
	_ = loginCmd.MarkFlagRequired("server-url")
	_ = loginCmd.MarkFlagRequired("cert-file")
}
