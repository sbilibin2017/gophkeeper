package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/spf13/cobra"
)

// NewRegisterCommand returns a cobra command to register a new user.
//
// This command accepts the following flags:
// - --username: the login name for the new user (required)
// - --password: the password for the new user (required)
// - --server-url: the URL of the server (required)
// - --client-public-key-file: the path to the client's public key PEM file (required)
//
// Based on the URL scheme, it determines whether to use HTTP or gRPC for the registration request.
func NewRegisterCommand() *cobra.Command {
	var username, password, serverURL, clientPubKeyFile string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			pubKey, err := os.ReadFile(clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to read public key: %w", err)
			}

			req := &models.AuthRequest{
				Login:     username,
				Password:  password,
				PublicKey: string(pubKey),
			}

			schemeType := scheme.GetSchemeFromURL(serverURL)

			switch schemeType {
			case scheme.HTTP, scheme.HTTPS:
				db, err := db.New("sqlite", "client.db",
					db.WithMaxOpenConns(10),
					db.WithMaxIdleConns(5),
					db.WithConnMaxLifetime(30*time.Minute),
				)
				if err != nil {
					return fmt.Errorf("failed to connect to DB: %w", err)
				}
				defer db.Close()

				client, err := http.New(serverURL)
				if err != nil {
					return err
				}

				err = repositories.CreateEncryptedSecretsTable(ctx, db)
				if err != nil {
					return err
				}

				auth := facades.NewAuthHTTPFacade(client)

				resp, err := auth.Register(ctx, req)
				if err != nil {
					return err
				}

				cmd.Println(resp.Token)

			case scheme.GRPC:
				db, err := db.New("sqlite", "client.db",
					db.WithMaxOpenConns(10),
					db.WithMaxIdleConns(5),
					db.WithConnMaxLifetime(30*time.Minute),
				)
				if err != nil {
					return fmt.Errorf("failed to connect to DB: %w", err)
				}
				defer db.Close()

				client, err := grpc.New(serverURL)
				if err != nil {
					return err
				}

				err = repositories.CreateEncryptedSecretsTable(ctx, db)
				if err != nil {
					return err
				}

				auth := facades.NewAuthGRPCFacade(client)

				resp, err := auth.Register(ctx, req)
				if err != nil {
					return err
				}

				cmd.Println(resp.Token)

			default:
				return fmt.Errorf("unsupported scheme: %s", schemeType)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username (required)")
	cmd.Flags().StringVar(&password, "password", "", "Password (required)")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL (required)")
	cmd.Flags().StringVar(&clientPubKeyFile, "client-public-key-file", "", "Path to client public key file (required)")

	return cmd
}

// NewLoginCommand returns a cobra command for user login.
//
// This command accepts:
// - --username / -u: login name (required)
// - --password / -p: password (required)
// - --server-url: server URL to authenticate against (required)
// - --client-public-key-file: path to the client's public key PEM file (required)
//
// It supports both HTTP and gRPC authentication depending on the URL scheme.
func NewLoginCommand() *cobra.Command {
	var username, password, serverURL, clientPubKeyFile string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login as an existing user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			pubKey, err := os.ReadFile(clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to read public key: %w", err)
			}

			req := &models.AuthRequest{
				Login:     username,
				Password:  password,
				PublicKey: string(pubKey),
			}

			schemeType := scheme.GetSchemeFromURL(serverURL)

			switch schemeType {
			case scheme.HTTP, scheme.HTTPS:
				dbConn, err := db.New("sqlite", "client.db",
					db.WithMaxOpenConns(10),
					db.WithMaxIdleConns(5),
					db.WithConnMaxLifetime(30*time.Minute),
				)
				if err != nil {
					return fmt.Errorf("failed to connect to DB: %w", err)
				}
				defer dbConn.Close()

				if err := repositories.CreateEncryptedSecretsTable(ctx, dbConn); err != nil {
					return err
				}

				client, err := http.New(serverURL)
				if err != nil {
					return fmt.Errorf("failed to create HTTP client: %w", err)
				}
				auth := facades.NewAuthHTTPFacade(client)

				resp, err := auth.Login(ctx, req)
				if err != nil {
					return fmt.Errorf("HTTP login failed: %w", err)
				}

				cmd.Println(resp.Token)

			case scheme.GRPC:
				dbConn, err := db.New("sqlite", "client.db",
					db.WithMaxOpenConns(10),
					db.WithMaxIdleConns(5),
					db.WithConnMaxLifetime(30*time.Minute),
				)
				if err != nil {
					return fmt.Errorf("failed to connect to DB: %w", err)
				}
				defer dbConn.Close()

				if err := repositories.CreateEncryptedSecretsTable(ctx, dbConn); err != nil {
					return err
				}

				client, err := grpc.New(serverURL)
				if err != nil {
					return fmt.Errorf("failed to create gRPC client: %w", err)
				}
				defer client.Close()

				auth := facades.NewAuthGRPCFacade(client)

				resp, err := auth.Login(ctx, req)
				if err != nil {
					return fmt.Errorf("gRPC login failed: %w", err)
				}

				cmd.Println(resp.Token)

			default:
				return fmt.Errorf("unsupported scheme: %s", schemeType)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username (required)")
	cmd.Flags().StringVar(&password, "password", "", "Password (required)")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL (required)")
	cmd.Flags().StringVar(&clientPubKeyFile, "client-public-key-file", "", "Path to client public key file (required)")

	return cmd
}

// NewLogoutCommand returns a cobra command for logging out the current user.
//
// It requires:
// - --server-url: server to logout from (required)
// - --token: the authentication token for the session (required)
//
// This command drops the local encrypted secrets table and informs the server to invalidate the token.
func NewLogoutCommand() *cobra.Command {
	var serverURL, token string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			dbConn, err := db.New("sqlite", "client.db",
				db.WithMaxOpenConns(10),
				db.WithMaxIdleConns(5),
				db.WithConnMaxLifetime(30*time.Minute),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to DB: %w", err)
			}
			defer dbConn.Close()

			if err := repositories.DropEncryptedSecretsTable(ctx, dbConn); err != nil {
				return fmt.Errorf("failed to drop encrypted secrets table: %w", err)
			}

			schemeType := scheme.GetSchemeFromURL(serverURL)

			switch schemeType {
			case scheme.HTTP, scheme.HTTPS:
				client, err := http.New(serverURL, http.WithAuthToken(token))
				if err != nil {
					return fmt.Errorf("failed to create HTTP client: %w", err)
				}
				auth := facades.NewAuthHTTPFacade(client)
				if err := auth.Logout(ctx); err != nil {
					return fmt.Errorf("HTTP logout failed: %w", err)
				}
				cmd.Println("Logout successful")

			case scheme.GRPC:
				client, err := grpc.New(serverURL, grpc.WithAuthToken(token))
				if err != nil {
					return fmt.Errorf("failed to create gRPC client: %w", err)
				}
				defer client.Close()

				auth := facades.NewAuthGRPCFacade(client)
				if err := auth.Logout(ctx); err != nil {
					return fmt.Errorf("gRPC logout failed: %w", err)
				}
				cmd.Println("Logout successful")

			default:
				return fmt.Errorf("unsupported scheme: %s", schemeType)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL (required)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token (required)")

	return cmd
}
