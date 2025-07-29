package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sbilibin2017/gophkeeper/inernal/apps"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	"github.com/sbilibin2017/gophkeeper/inernal/usecases"
	"github.com/spf13/cobra"
)

// NewClientCommand returns the root cobra command for the GophKeeper CLI.
func NewClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper CLI - secure personal data manager",
		Long: `GophKeeper is a secure personal data manager that allows you to store 
various secret types (bank cards, login credentials, binary data, etc.). 
Use this CLI to interact with the secure backend for registration, login, adding secrets, listing, and syncing.`,
	}
	return cmd
}

// NewClientInfoCommand returns a cobra command that prints version and build info.
func NewClientInfoCommand(buildVersion, buildDate string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Print build version and build date information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GophKeeper CLI version: %s\n", buildVersion)
			fmt.Printf("Build date: %s\n", buildDate)
		},
	}
	return cmd
}

// NewClientRegisterCommand creates and returns a Cobra command for user registration.
func NewClientRegisterCommand() *cobra.Command {
	var (
		serverURL        string
		clientPubKeyFile string

		username string
		password string
	)

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			schemeType := scheme.GetSchemeFromURL(serverURL)

			var app *usecases.ClientRegisterUsecase
			var err error

			switch schemeType {
			case scheme.GRPC:
				app, err = apps.NewClientRegisterGRPCApp(
					serverURL,
					"sqlite",
					"client.db",
					"../../../migrations",
				)
				if err != nil {
					return fmt.Errorf("failed to initialize GRPC register app: %w", err)
				}

			case scheme.HTTP, scheme.HTTPS:
				app, err = apps.NewClientRegisterHTTPApp(serverURL,
					"sqlite",
					"client.db",
					"../../../migrations",
				)
				if err != nil {
					return fmt.Errorf("failed to initialize HTTP register app: %w", err)
				}

			default:
				return fmt.Errorf("unsupported scheme %s", schemeType)
			}

			req := models.AuthRegisterRequest{
				Username: username,
				Password: password,
			}

			resp, err := app.Execute(ctx, req)
			if err != nil {
				return fmt.Errorf("registration failed: %w", err)
			}

			cmd.Println("Registration successful. Access Token:")
			cmd.Println(resp.Token)

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL of the authentication server (grpc:// or http://)")
	cmd.Flags().StringVar(&clientPubKeyFile, "pubkey-file", "", "Path to client public key PEM file (optional)")

	cmd.Flags().StringVar(&username, "username", "", "Username for registration")
	cmd.Flags().StringVar(&password, "password", "", "Password for registration")

	return cmd
}

// NewLoginCommand creates the login CLI command.
func NewClientLoginCommand() *cobra.Command {
	var (
		serverURL string
		username  string
		password  string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login an existing user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			schemeType := scheme.GetSchemeFromURL(serverURL)

			var app *usecases.ClientLoginUsecase
			var err error

			switch schemeType {
			case scheme.GRPC:
				app, err = apps.NewClientLoginGRPCApp(serverURL)
				if err != nil {
					return fmt.Errorf("failed to initialize GRPC login app: %w", err)
				}

			case scheme.HTTP, scheme.HTTPS:
				app, err = apps.NewClientLoginHTTPApp(serverURL)
				if err != nil {
					return fmt.Errorf("failed to initialize HTTP login app: %w", err)
				}

			default:
				return fmt.Errorf("unsupported scheme %s", schemeType)
			}

			req := models.AuthLoginRequest{
				Username: username,
				Password: password,
			}

			resp, err := app.Execute(ctx, req)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			cmd.Println(resp.Token)

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL of the authentication server (grpc:// or http://)")

	cmd.Flags().StringVar(&username, "username", "", "Username for login")
	cmd.Flags().StringVar(&password, "password", "", "Password for login")

	return cmd
}

// NewClientAddBankcardCommand creates and returns a Cobra command for add bank card secret.
func NewClientAddBankcardCommand() *cobra.Command {
	var (
		clientPubKeyFile string
		token            string

		secretName string
		number     string
		owner      string
		exp        string
		cvv        string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "add-bankcard",
		Short: "Add a new encrypted bankcard secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			app, err := apps.NewClientBankcardAddApp(
				"sqlite",
				"client.db",
				clientPubKeyFile,
			)
			if err != nil {
				return fmt.Errorf("failed to initialize bankcard add app: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			req := &models.BankcardAddRequest{
				Token:      token,
				SecretName: secretName,
				Number:     number,
				Owner:      owner,
				Exp:        exp,
				CVV:        cvv,
				Meta:       metaPtr,
			}

			if err := app.Execute(ctx, req); err != nil {
				return fmt.Errorf("failed to add bankcard: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&clientPubKeyFile, "pubkey-file", "", "Path to server public key PEM file")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret")
	cmd.Flags().StringVar(&number, "number", "", "Card number")
	cmd.Flags().StringVar(&owner, "owner", "", "Cardholder name")
	cmd.Flags().StringVar(&exp, "exp", "", "Expiration date (MM/YY)")
	cmd.Flags().StringVar(&cvv, "cvv", "", "Card CVV")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}

// NewClientAddBinaryCommand creates and returns a Cobra command for add binary secret.
func NewClientAddBinaryCommand() *cobra.Command {
	var (
		clientPubKeyFile string
		token            string
		secretName       string
		filePath         string
		meta             string
	)

	cmd := &cobra.Command{
		Use:   "add-binary",
		Short: "Add a new encrypted binary secret from a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			app, err := apps.NewClientBinaryAddApp("sqlite", "client.db", clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to initialize binary add app: %w", err)
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			req := &models.BinaryAddRequest{
				Token:      token,
				SecretName: secretName,
				Data:       data,
				Meta:       metaPtr,
			}

			if err := app.Execute(ctx, req); err != nil {
				return fmt.Errorf("failed to add binary secret: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&clientPubKeyFile, "pubkey-file", "", "Path to public key PEM file")
	cmd.Flags().StringVar(&token, "token", "", "Auth token")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to binary file to encrypt")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}

// NewClientAddTextCommand creates and returns a Cobra command for add text secret.
func NewClientAddTextCommand() *cobra.Command {
	var (
		clientPubKeyFile string
		token            string
		secretName       string
		data             string
		meta             string
	)

	cmd := &cobra.Command{
		Use:   "add-text",
		Short: "Add a new encrypted text secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			app, err := apps.NewClientTextAddApp("sqlite", "client.db", clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to initialize text add app: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			req := &models.TextAddRequest{
				Token:      token,
				SecretName: secretName,
				Data:       data,
				Meta:       metaPtr,
			}

			if err := app.Execute(ctx, req); err != nil {
				return fmt.Errorf("failed to add text secret: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&clientPubKeyFile, "pubkey-file", "", "Path to public key PEM file")
	cmd.Flags().StringVar(&token, "token", "", "Auth token")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret")
	cmd.Flags().StringVar(&data, "data", "", "Text data to encrypt")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}

// NewClientAddUserCommand creates and returns a Cobra command for add user secret.
func NewClientAddUserCommand() *cobra.Command {
	var (
		clientPubKeyFile string
		token            string
		secretName       string
		username         string
		password         string
		meta             string
	)

	cmd := &cobra.Command{
		Use:   "add-user",
		Short: "Add a new encrypted user/password secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			app, err := apps.NewClientUserAddApp("sqlite", "client.db", clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to initialize user add app: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			req := &models.UserAddRequest{
				Token:      token,
				SecretName: secretName,
				Username:   username,
				Password:   password,
				Meta:       metaPtr,
			}

			if err := app.Execute(ctx, req); err != nil {
				return fmt.Errorf("failed to add user secret: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&clientPubKeyFile, "pubkey-file", "", "Path to public key PEM file")
	cmd.Flags().StringVar(&token, "token", "", "Auth token")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret")
	cmd.Flags().StringVar(&username, "username", "", "Username to store")
	cmd.Flags().StringVar(&password, "password", "", "Password to store")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}

// NewListCommand returns a cobra command that lists all secrets using either HTTP or gRPC
func NewClientListCommand() *cobra.Command {
	var (
		serverURL         string
		token             string
		clientPrivKeyFile string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all secrets from server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			schemeType := scheme.GetSchemeFromURL(serverURL)

			var app *usecases.ClientListUsecase
			var err error

			switch schemeType {
			case scheme.HTTP, scheme.HTTPS:
				app, err = apps.NewClientListHTTPApp(serverURL, clientPrivKeyFile) // private key here
				if err != nil {
					return fmt.Errorf("failed to initialize HTTP list app: %w", err)
				}
			case scheme.GRPC:
				app, err = apps.NewClientListGRPCApp(serverURL, clientPrivKeyFile) // private key here
				if err != nil {
					return fmt.Errorf("failed to initialize gRPC list app: %w", err)
				}
			default:
				return fmt.Errorf("unsupported scheme %s", schemeType)
			}

			req := &models.SecretListRequest{
				Token: token,
			}

			bankcards, users, texts, binaries, err := app.Execute(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to list secrets: %w", err)
			}

			if len(bankcards) == 0 && len(users) == 0 && len(texts) == 0 && len(binaries) == 0 {
				cmd.Println("No secrets found.")
				return nil
			}

			// After fetching secrets from app.Execute(...)
			for _, b := range bankcards {
				jsonBytes, err := json.MarshalIndent(b, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal bankcard to JSON: %w", err)
				}
				cmd.Println(string(jsonBytes))
			}

			for _, u := range users {
				jsonBytes, err := json.MarshalIndent(u, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal user to JSON: %w", err)
				}
				cmd.Println(string(jsonBytes))
			}

			for _, t := range texts {
				jsonBytes, err := json.MarshalIndent(t, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal text to JSON: %w", err)
				}
				cmd.Println(string(jsonBytes))
			}

			for _, b := range binaries {
				jsonBytes, err := json.MarshalIndent(b, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal binary to JSON: %w", err)
				}
				cmd.Println(string(jsonBytes))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL of the secret server (http(s):// or grpc://)")
	cmd.Flags().StringVar(&clientPrivKeyFile, "privkey-file", "", "Path to client private key PEM file")
	cmd.Flags().StringVar(&token, "token", "", "Access token for authentication")

	cmd.MarkFlagRequired("server-url")
	cmd.MarkFlagRequired("privkey-file")
	cmd.MarkFlagRequired("token")

	return cmd
}

// NewClientSyncCommand returns a cobra command that syncs server with client.
func NewClientSyncCommand() *cobra.Command {
	var (
		serverURL         string
		token             string
		clientPrivKeyFile string
		mode              string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize client secrets to the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			schemeType := scheme.GetSchemeFromURL(serverURL)

			switch mode {
			case models.SyncModeServer:
				serverSync := apps.NewServerSyncApp()
				if err := serverSync.Sync(ctx); err != nil {
					return fmt.Errorf("failed to sync (server mode): %w", err)
				}

			case models.SyncModeClient:
				var syncUsecase interface {
					Sync(ctx context.Context, token string) error
				}
				var err error

				switch schemeType {
				case scheme.HTTP, scheme.HTTPS:
					syncUsecase, err = apps.NewClientSyncHTTPApp("sqlite", "client.db", serverURL)
					if err != nil {
						return fmt.Errorf("failed to initialize HTTP sync app: %w", err)
					}
				case scheme.GRPC:
					syncUsecase, err = apps.NewClientSyncGRPCApp("sqlite", "client.db", serverURL)
					if err != nil {
						return fmt.Errorf("failed to initialize gRPC sync app: %w", err)
					}
				default:
					return fmt.Errorf("unsupported scheme %s", schemeType)
				}

				if err := syncUsecase.Sync(ctx, token); err != nil {
					return fmt.Errorf("failed to sync secrets: %w", err)
				}

			case models.SyncModeInteractive:
				var interactiveUsecase interface {
					Sync(ctx context.Context, reader io.Reader, token string) error
				}
				var err error

				switch schemeType {
				case scheme.HTTP, scheme.HTTPS:
					interactiveUsecase, err = apps.NewSyncInteractiveHTTPApp("sqlite", "client.db", serverURL, clientPrivKeyFile)
					if err != nil {
						return fmt.Errorf("failed to initialize interactive HTTP sync app: %w", err)
					}
				case scheme.GRPC:
					interactiveUsecase, err = apps.NewSyncInteractiveGRPCApp("sqlite", "client.db", serverURL, clientPrivKeyFile)
					if err != nil {
						return fmt.Errorf("failed to initialize interactive gRPC sync app: %w", err)
					}
				default:
					return fmt.Errorf("unsupported scheme %s", schemeType)
				}

				if err := interactiveUsecase.Sync(ctx, cmd.InOrStdin(), token); err != nil {
					return fmt.Errorf("failed to sync interactively: %w", err)
				}

			default:
				return fmt.Errorf("unsupported mode %s", mode)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL of the secret server (http(s):// or grpc://)")
	cmd.Flags().StringVar(&clientPrivKeyFile, "privkey-file", "", "Path to client private key PEM file (required for interactive mode)")
	cmd.Flags().StringVar(&token, "token", "", "Access token for authentication")
	cmd.Flags().StringVar(&mode, "mode", models.SyncModeClient, "Sync mode: server, client, or interactive")

	return cmd
}
