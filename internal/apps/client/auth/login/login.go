package login

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/facades/auth"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/user"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/spf13/cobra"
)

// RegisterCommand adds the "login" subcommand to the provided Cobra root command.
// This command allows a user to authenticate against an HTTP or gRPC service.
//
// Parameters:
//   - root: The root Cobra command to which the "login" subcommand will be added.
//   - runHTTPFunc: A function that performs HTTP-based authentication using context, username, and password.
//   - runGRPCFunc: A function that performs gRPC-based authentication using context, username, and password.
//
// The subcommand accepts the following flags:
//
//	--username         Required: The username to login with.
//	--password         Required: The password for the specified user.
//	--auth-url         Required: The URL of the authentication service (must begin with grpc://, http://, or https://).
//	--tls-client-cert  Required: Path to the client TLS certificate file.
//	--tls-client-key   Required: Path to the client TLS key file.
//
// Example usage:
//
//	gophkeeper login \
//	  --username alice \
//	  --password S3cr3tPass! \
//	  --auth-url https://example.com \
//	  --tls-client-cert cert.pem \
//	  --tls-client-key key.pem
func RegisterCommand(
	root *cobra.Command,
	runHTTPFunc func(ctx context.Context, username, password string) (*models.AuthResponse, error),
	runGRPCFunc func(ctx context.Context, username, password string) (*models.AuthResponse, error),
) {
	var (
		username    string
		password    string
		authURL     string
		tlsCertFile string
		tlsKeyFile  string
	)

	cmd := &cobra.Command{
		Use:     "login",
		Short:   "Login a user",
		Long:    "Authenticate a user by username, password, and authentication details to obtain a session token.",
		Example: `gophkeeper login --username alice --password "S3cr3tPass!" --auth-url https://example.com --tls-client-cert cert.pem --tls-client-key key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var resp *models.AuthResponse
			var err error

			switch {
			case strings.HasPrefix(authURL, "grpc://"):
				resp, err = runGRPCFunc(ctx, username, password)
			case strings.HasPrefix(authURL, "http://"), strings.HasPrefix(authURL, "https://"):
				resp, err = runHTTPFunc(ctx, username, password)
			default:
				return fmt.Errorf("unsupported auth URL scheme, must start with grpc://, http:// or https://")
			}

			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			cmd.Println(resp.Token)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username for login")
	cmd.Flags().StringVar(&password, "password", "", "Password for login")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication server URL")
	cmd.Flags().StringVar(&tlsCertFile, "tls-client-cert", "", "Path to TLS client certificate file")
	cmd.Flags().StringVar(&tlsKeyFile, "tls-client-key", "", "Path to TLS client key file")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("auth-url")
	cmd.MarkFlagRequired("tls-client-cert")
	cmd.MarkFlagRequired("tls-client-key")

	root.AddCommand(cmd)
}

// NewRunGRPC returns a closure function that performs user authentication
// via gRPC using the provided server address and TLS certificate files.
//
// Parameters:
//   - authURL: The full gRPC service address (e.g., "grpc://localhost:50051").
//   - tlsCertFile: Path to the TLS certificate file for the client.
//   - tlsKeyFile: Path to the TLS key file for the client.
//
// The returned function accepts:
//   - ctx: Context for timeout and cancellation control.
//   - username: The username of the user attempting to authenticate.
//   - password: The user's password.
//
// Returns:
//   - *models.AuthResponse: Contains access and refresh tokens if authentication succeeds.
//   - error: Any error encountered during DB setup, connection, or login process.
func NewRunGRPC(authURL, tlsCertFile, tlsKeyFile string) func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
	return func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
		dbConn, err := db.NewDB("sqlite", "client.db")
		if err != nil {
			return nil, fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer dbConn.Close()

		if err := bankcard.CreateClientTable(ctx, dbConn); err != nil {
			return nil, fmt.Errorf("failed to create bankcard table: %w", err)
		}
		if err := text.CreateClientTable(ctx, dbConn); err != nil {
			return nil, fmt.Errorf("failed to create text table: %w", err)
		}
		if err := binary.CreateClientTable(ctx, dbConn); err != nil {
			return nil, fmt.Errorf("failed to create binary table: %w", err)
		}
		if err := user.CreateClientTable(ctx, dbConn); err != nil {
			return nil, fmt.Errorf("failed to create user table: %w", err)
		}

		grpcConn, err := grpc.New(
			strings.TrimPrefix(authURL, "grpc://"),
			grpc.WithTLSCert(grpc.TLSCert{CertFile: tlsCertFile, KeyFile: tlsKeyFile}),
			grpc.WithRetryPolicy(grpc.RetryPolicy{
				Count:   3,
				Wait:    2 * time.Second,
				MaxWait: 10 * time.Second,
			}),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
		}
		defer grpcConn.Close()

		grpcClient := pb.NewAuthServiceClient(grpcConn)
		facade := auth.NewLoginGRPCFacade(grpcClient)

		authReq := &models.AuthRequest{
			Username: username,
			Password: password,
		}

		resp, err := facade.Login(ctx, authReq)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

// NewRunHTTP returns a closure function that performs user authentication
// via an HTTP client using the specified authentication server URL and TLS certificate files.
//
// Parameters:
//   - authURL: The full URL of the authentication HTTP server (e.g., "https://auth.example.com").
//   - tlsCertFile: Path to the TLS certificate file for the client.
//   - tlsKeyFile: Path to the TLS key file for the client.
//
// The returned function accepts:
//   - ctx: Context for managing request lifecycle (e.g., timeouts, cancellations).
//   - username: The username of the user attempting to authenticate.
//   - password: The user's password.
//
// Returns:
//   - *models.AuthResponse: Contains access and refresh tokens if authentication is successful.
//   - error: Any error encountered during database setup, HTTP client creation, or login process.
func NewRunHTTP(authURL, tlsCertFile, tlsKeyFile string) func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
	return func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
		conn, err := db.NewDB("sqlite", "client.db")
		if err != nil {
			return nil, fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer conn.Close()

		if err := bankcard.CreateClientTable(ctx, conn); err != nil {
			return nil, err
		}
		if err := text.CreateClientTable(ctx, conn); err != nil {
			return nil, err
		}
		if err := binary.CreateClientTable(ctx, conn); err != nil {
			return nil, err
		}
		if err := user.CreateClientTable(ctx, conn); err != nil {
			return nil, err
		}

		client, err := http.New(
			authURL,
			http.WithTLSCert(http.TLSCert{CertFile: tlsCertFile, KeyFile: tlsKeyFile}),
			http.WithRetryPolicy(http.RetryPolicy{Count: 3, Wait: 2 * time.Second}),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP client: %w", err)
		}

		facade := auth.NewLoginHTTPFacade(client)

		authReq := &models.AuthRequest{
			Username: username,
			Password: password,
		}

		resp, err := facade.Login(ctx, authReq)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}
