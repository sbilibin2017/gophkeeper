package logout

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/facades/auth"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/user"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/spf13/cobra"
)

// RegisterLogoutCommand adds the 'logout' command to the provided root Cobra command.
func RegisterLogoutCommand(
	root *cobra.Command,
	runLogoutHTTP func(ctx context.Context, token string) error,
	runLogoutGRPC func(ctx context.Context, token string) error,
) {
	var (
		authURL     string
		tlsCertFile string
		tlsKeyFile  string
		token       string
	)

	cmd := &cobra.Command{
		Use:     "logout",
		Short:   "Logout the current user",
		Long:    "Logout the current user and invalidate the session token.",
		Example: `gophkeeper logout --auth-url https://example.com --token your-token --tls-client-cert cert.pem --tls-client-key key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var err error
			switch {
			case strings.HasPrefix(authURL, "grpc://"):
				err = runLogoutGRPC(ctx, token)
			case strings.HasPrefix(authURL, "http://"), strings.HasPrefix(authURL, "https://"):
				err = runLogoutHTTP(ctx, token)
			default:
				return fmt.Errorf("unsupported auth URL scheme, must start with grpc://, http:// or https://")
			}

			if err != nil {
				return fmt.Errorf("logout failed: %w", err)
			}

			cmd.Println("Logout successful.")
			return nil
		},
	}

	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication server URL")
	cmd.Flags().StringVar(&tlsCertFile, "tls-client-cert", "", "Path to TLS client certificate file")
	cmd.Flags().StringVar(&tlsKeyFile, "tls-client-key", "", "Path to TLS client key file")
	cmd.Flags().StringVar(&token, "token", "", "Session token to logout")

	cmd.MarkFlagRequired("auth-url")
	cmd.MarkFlagRequired("tls-client-cert")
	cmd.MarkFlagRequired("tls-client-key")
	cmd.MarkFlagRequired("token")

	root.AddCommand(cmd)
}

// NewRunGRPC returns a closure that performs logout via gRPC.
func NewRunGRPC(authURL, tlsCertFile, tlsKeyFile string) func(ctx context.Context, token string) error {
	return func(ctx context.Context, token string) error {
		// Setup DB connection
		dbConn, err := db.NewDB("sqlite", "client.db")
		if err != nil {
			return fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer dbConn.Close()

		conn, err := grpc.New(
			authURL,
			grpc.WithTLSCert(grpc.TLSCert{CertFile: tlsCertFile, KeyFile: tlsKeyFile}),
			grpc.WithAuthToken(token),
			grpc.WithRetryPolicy(grpc.RetryPolicy{
				Count:   3,
				Wait:    2 * time.Second,
				MaxWait: 10 * time.Second,
			}),
		)
		if err != nil {
			return fmt.Errorf("failed to create gRPC connection: %w", err)
		}
		defer conn.Close()

		client := pb.NewAuthServiceClient(conn)
		facade := auth.NewLogoutGRPCFacade(client)

		if err := facade.Logout(ctx); err != nil {
			return err
		}

		bankcard.DropClientTable(ctx, dbConn)
		text.DropClientTable(ctx, dbConn)
		binary.DropClientTable(ctx, dbConn)
		user.DropClientTable(ctx, dbConn)

		return nil
	}
}

// NewRunHTTP returns a closure that performs logout via HTTP.
func NewRunHTTP(authURL, tlsCertFile, tlsKeyFile string) func(ctx context.Context, token string) error {
	return func(ctx context.Context, token string) error {
		// Setup DB connection
		dbConn, err := db.NewDB("sqlite", "client.db")
		if err != nil {
			return fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer dbConn.Close()

		// Create HTTP client with TLS cert and retry policy
		client, err := http.New(
			authURL,
			http.WithTLSCert(http.TLSCert{CertFile: tlsCertFile, KeyFile: tlsKeyFile}),
			http.WithRetryPolicy(http.RetryPolicy{
				Count: 3,
				Wait:  2 * time.Second,
			}),
			http.WithAuthToken(token),
		)
		if err != nil {
			return fmt.Errorf("failed to create HTTP client: %w", err)
		}

		facade := auth.NewLogoutHTTPFacade(client)

		if err := facade.Logout(ctx); err != nil {
			return err
		}

		bankcard.DropClientTable(ctx, dbConn)
		text.DropClientTable(ctx, dbConn)
		binary.DropClientTable(ctx, dbConn)
		user.DropClientTable(ctx, dbConn)

		return nil
	}
}
