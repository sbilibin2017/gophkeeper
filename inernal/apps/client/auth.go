package client

import (
	"fmt"
	"os"

	"github.com/pressly/goose"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/db"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/inernal/facades"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	"github.com/spf13/cobra"
)

// NewRegisterCommand returns a cobra command stub that registers a new user.
// Currently, the command is not implemented and prints a placeholder message.
// Adds flags: --server-url, --client-pub-key-file, --username, --password.
func NewAuthRegisterCommand() *cobra.Command {
	var serverURL string
	var clientPubKeyFile string
	var username string
	var password string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			pubKey, err := os.ReadFile(clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to read public key: %w", err)
			}

			db, err := db.New("sqlite", "client.db")
			if err != nil {
				return fmt.Errorf("failed to connect to DB: %w", err)
			}
			defer db.Close()

			req := &models.UserRegisterRequest{
				Username:         username,
				Password:         password,
				ClientPubKeyFile: string(pubKey),
			}

			schemeType := scheme.GetSchemeFromURL(serverURL)

			switch schemeType {
			case scheme.HTTP, scheme.HTTPS:
				client, err := http.New(serverURL)
				if err != nil {
					return err
				}

				if err := goose.Up(db.DB, "../../../migrations"); err != nil {
					return err
				}

				auth, err := facades.NewAuthHTTPFacade(client)
				if err != nil {
					return err
				}

				resp, err := auth.Register(ctx, req)
				if err != nil {
					return err
				}

				cmd.Println(resp.Token)

			case scheme.GRPC:
				client, err := grpc.New(serverURL)
				if err != nil {
					return err
				}

				auth, err := facades.NewAuthGRPCFacade(client)
				if err != nil {
					return err
				}

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

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL of the server")
	cmd.Flags().StringVar(&clientPubKeyFile, "client-pub-key-file", "", "Path to client public key file")

	cmd.Flags().StringVar(&username, "username", "", "Username for registration")
	cmd.Flags().StringVar(&password, "password", "", "Password for registration")

	return cmd
}

// NewLoginCommand returns a cobra command stub that logs in an existing user.
// Currently, the command is not implemented and prints a placeholder message.
// Adds flags: --server-url, --username, --password.
// NewAuthLoginCommand returns a cobra command stub that logs in an existing user.
// It supports HTTP and gRPC schemes based on the server URL.
// Adds flags: --server-url, --client-pub-key-file, --username, --password.
func NewAuthLoginCommand() *cobra.Command {
	var serverURL string
	var username string
	var password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			db, err := db.New("sqlite", "client.db")
			if err != nil {
				return fmt.Errorf("failed to connect to DB: %w", err)
			}
			defer db.Close()

			req := &models.UserLoginRequest{
				Username: username,
				Password: password,
			}

			schemeType := scheme.GetSchemeFromURL(serverURL)

			switch schemeType {
			case scheme.HTTP, scheme.HTTPS:
				client, err := http.New(serverURL)
				if err != nil {
					return err
				}

				if err := goose.Up(db.DB, "../../../migrations"); err != nil {
					return err
				}

				auth, err := facades.NewAuthHTTPFacade(client)
				if err != nil {
					return err
				}

				resp, err := auth.Login(ctx, req)
				if err != nil {
					return err
				}

				cmd.Println(resp.Token)

			case scheme.GRPC:
				client, err := grpc.New(serverURL)
				if err != nil {
					return err
				}

				auth, err := facades.NewAuthGRPCFacade(client)
				if err != nil {
					return err
				}

				resp, err := auth.Login(ctx, req)
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

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL of the server")

	cmd.Flags().StringVar(&username, "username", "", "Username for login")
	cmd.Flags().StringVar(&password, "password", "", "Password for login")

	return cmd
}
