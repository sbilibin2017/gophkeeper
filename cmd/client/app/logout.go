package app

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout the current user",
	Long:  "Logout the current user using TLS client authentication",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		serverURL, _ := cmd.Flags().GetString("server-url")
		certFile, _ := cmd.Flags().GetString("cert-file")
		certKey, _ := cmd.Flags().GetString("cert-key")

		schemeType := scheme.GetSchemeFromURL(serverURL)

		switch schemeType {
		case scheme.HTTP, scheme.HTTPS:
			if err := runLogoutHTTP(ctx, serverURL, certFile, certKey); err != nil {
				return err
			}
			cmd.Println("Logout successful")
		case scheme.GRPC:
			if err := runLogoutGRPC(ctx, serverURL, certFile, certKey); err != nil {
				return err
			}
			cmd.Println("Logout successful")
		default:
			return fmt.Errorf("unsupported scheme: %s", schemeType)
		}

		return nil
	},
}

func runLogoutHTTP(ctx context.Context, serverURL, certFile, certKey string) error {
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
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	authFacade := facades.NewAuthHTTPFacade(client)
	if err := authFacade.Logout(ctx); err != nil {
		return err
	}

	return nil
}

func runLogoutGRPC(ctx context.Context, serverURL, certFile, certKey string) error {
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
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	authFacade := facades.NewAuthGRPCFacade(conn)
	if err := authFacade.Logout(ctx); err != nil {
		return err
	}

	return nil
}

func init() {
	logoutCmd.Flags().String("server-url", "", "Server URL (required)")
	logoutCmd.Flags().String("cert-file", "", "Path to client certificate file (required)")
	logoutCmd.Flags().String("cert-key", "", "Path to client private key file (required)")

	_ = logoutCmd.MarkFlagRequired("server-url")
	_ = logoutCmd.MarkFlagRequired("cert-file")
	_ = logoutCmd.MarkFlagRequired("cert-key")
}
