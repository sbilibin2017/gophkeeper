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

// RegisterCommand adds the "login" subcommand to the root command.
//
// Flags:
//
//	--username
//	--password
//	--auth-url
//	--tls-client-cert
//	--tls-client-key
//
// Example:
//
//	gophkeeper login \
//	  --username alice \
//	  --password S3cr3tPass! \
//	  --auth-url https://example.com \
//	  --tls-client-cert cert.pem \
//	  --tls-client-key key.pem
func RegisterLoginCommand(
	root *cobra.Command,
	runHTTPFunc func(ctx context.Context, authURL, tlsCertFile, tlsKeyFile, username, password string) (*models.AuthResponse, error),
	runGRPCFunc func(ctx context.Context, authURL, tlsCertFile, tlsKeyFile, username, password string) (*models.AuthResponse, error),
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
				resp, err = runGRPCFunc(ctx, authURL, tlsCertFile, tlsKeyFile, username, password)
			case strings.HasPrefix(authURL, "http://"), strings.HasPrefix(authURL, "https://"):
				resp, err = runHTTPFunc(ctx, authURL, tlsCertFile, tlsKeyFile, username, password)
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

// RunLoginGRPC performs user login via gRPC.
func RunLoginGRPC(ctx context.Context, authURL, tlsCertFile, tlsKeyFile, username, password string) (*models.AuthResponse, error) {
	dbConn, err := db.NewDB("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	if err := bankcard.CreateClientTable(ctx, dbConn); err != nil {
		return nil, err
	}
	if err := text.CreateClientTable(ctx, dbConn); err != nil {
		return nil, err
	}
	if err := binary.CreateClientTable(ctx, dbConn); err != nil {
		return nil, err
	}
	if err := user.CreateClientTable(ctx, dbConn); err != nil {
		return nil, err
	}

	conn, err := grpc.New(
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
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	facade := auth.NewLoginGRPCFacade(client)

	return facade.Login(ctx, &models.AuthRequest{
		Username: username,
		Password: password,
	})
}

// RunLoginHTTP performs user login via HTTP.
func RunLoginHTTP(ctx context.Context, authURL, tlsCertFile, tlsKeyFile, username, password string) (*models.AuthResponse, error) {
	dbConn, err := db.NewDB("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	if err := bankcard.CreateClientTable(ctx, dbConn); err != nil {
		return nil, err
	}
	if err := text.CreateClientTable(ctx, dbConn); err != nil {
		return nil, err
	}
	if err := binary.CreateClientTable(ctx, dbConn); err != nil {
		return nil, err
	}
	if err := user.CreateClientTable(ctx, dbConn); err != nil {
		return nil, err
	}

	client, err := http.New(
		authURL,
		http.WithTLSCert(http.TLSCert{CertFile: tlsCertFile, KeyFile: tlsKeyFile}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count: 3,
			Wait:  2 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	facade := auth.NewLoginHTTPFacade(client)

	return facade.Login(ctx, &models.AuthRequest{
		Username: username,
		Password: password,
	})
}
