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

	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("client-public-key-file")

	return cmd
}

func NewLoginCommand() *cobra.Command {
	var username, password, serverURL string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login as an existing user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			req := &models.AuthRequest{
				Login:    username,
				Password: password,
			}

			schemeType := scheme.GetSchemeFromURL(serverURL)

			switch schemeType {
			case scheme.HTTP, scheme.HTTPS:
				// Open DB connection & create tables before login
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
				// Open DB connection & create tables before login
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

	cmd.Flags().StringVarP(&username, "username", "u", "", "Username (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL (required)")

	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("server-url")

	return cmd
}

func NewLogoutCommand() *cobra.Command {
	var serverURL, token string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Open DB connection
			dbConn, err := db.New("sqlite", "client.db",
				db.WithMaxOpenConns(10),
				db.WithMaxIdleConns(5),
				db.WithConnMaxLifetime(30*time.Minute),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to DB: %w", err)
			}
			defer dbConn.Close()

			// Drop tables
			if err := repositories.DropEncryptedSecretsTable(ctx, dbConn); err != nil {
				return fmt.Errorf("failed to drop bankcards table: %w", err)
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

	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("token")

	return cmd
}
