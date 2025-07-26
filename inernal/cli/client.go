package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/sbilibin2017/gophkeeper/inernal/apps"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	"github.com/spf13/cobra"
)

// NewInfoCommand creates an "info" command that prints build info.
func NewClientInfoCommand(buildVersion, buildDate string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show build version and date",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Build Version: %s\n", buildVersion)
			fmt.Printf("Build Date: %s\n", buildDate)
		},
	}
	return cmd
}

// NewRootCommand returns the root cobra command for the GophKeeper CLI.
// It serves as the entry point for all subcommands and provides basic usage information.
func NewClientCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper is a secure personal data manager",
		Long:  "GophKeeper CLI lets you register, login, and logout users securely using TLS authentication.",
	}
}

// NewRegisterCommand creates the "register" command to register a new user.
// Flags:
// - server-url: URL of the server
// - client-pub-key-path: Path to client public key (required for registration)
// - username: Username for registration
// - password: Password for registration
func NewClientRegisterCommand() *cobra.Command {
	var (
		serverURL        string
		clientPubKeyFile string
		username         string
		password         string
	)

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			pubKey, err := os.ReadFile(clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to read public key file: %w", err)
			}

			app, err := apps.NewClientRegisterApp(serverURL)
			if err != nil {
				return fmt.Errorf("failed to initialize register app: %w", err)
			}

			req := &models.UserRegisterRequest{
				Username:         username,
				Password:         password,
				ClientPubKeyFile: string(pubKey),
			}

			resp, err := app.Execute(ctx, req)
			if err != nil {
				return fmt.Errorf("registration failed: %w", err)
			}

			cmd.Println("Registration successful. Token:")
			cmd.Println(resp.Token)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL of the server")
	cmd.Flags().StringVar(&clientPubKeyFile, "client-pub-key-path", "", "Path to client public key")
	cmd.Flags().StringVar(&username, "username", "", "Username for registration")
	cmd.Flags().StringVar(&password, "password", "", "Password for registration")

	return cmd
}

// NewLoginCommand creates the "login" command that allows a user to login.
// Flags:
// - server-url: URL of the server
// - client-pub-key-path: Path to client public key (currently unused in login)
// - username: Username for login
// - password: Password for login
func NewClientLoginCommand() *cobra.Command {
	var (
		serverURL string
		username  string
		password  string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			app, err := apps.NewClientLoginApp(serverURL)
			if err != nil {
				return fmt.Errorf("failed to initialize login app: %w", err)
			}

			req := &models.UserLoginRequest{
				Username: username,
				Password: password,
			}

			resp, err := app.Execute(ctx, req)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			cmd.Println("Login successful. Token:")
			cmd.Println(resp.Token)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL of the server")
	cmd.Flags().StringVar(&username, "username", "", "Username for login")
	cmd.Flags().StringVar(&password, "password", "", "Password for login")

	return cmd
}

// NewAddBankCardCommand creates the "add-bankcard" command to add a bank card secret.
// Flags:
// - client-pub-key-path: Path to client public key
// - token: JWT token for authentication (required)
// - name: Name of the secret (required)
// - number: Card number (required)
// - owner: Card owner's name (required)
// - exp: Card expiration date (MM/YY) (required)
// - cvv: Card CVV code (required)
// - meta: Optional metadata
func NewClientAddBankCardCommand() *cobra.Command {
	var (
		clientPubKeyFile string
		token            string
		secretName       string
		number           string
		owner            string
		exp              string
		cvv              string
		meta             string
	)

	cmd := &cobra.Command{
		Use:   "add-bankcard",
		Short: "Add a new bank card entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			app, err := apps.NewBankCardAddApp(clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to initialize add app: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			secret := &models.BankcardSecretAdd{
				SecretName:  secretName,
				SecretType:  models.SecretTypeBankCard,
				SecretOwner: token,
				Number:      number,
				Owner:       owner,
				Exp:         exp,
				CVV:         cvv,
				Meta:        metaPtr,
			}

			if err := app.Execute(ctx, secret, token); err != nil {
				return fmt.Errorf("failed to save secret: %w", err)
			}

			fmt.Println("Bank card secret saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&clientPubKeyFile, "client-pub-key-path", "", "Path to client public key")
	cmd.Flags().StringVar(&token, "token", "", "JWT token for authentication (required)")
	cmd.Flags().StringVar(&secretName, "name", "", "Name of the secret (required)")
	cmd.Flags().StringVar(&number, "number", "", "Card number (required)")
	cmd.Flags().StringVar(&owner, "owner", "", "Card owner's name (required)")
	cmd.Flags().StringVar(&exp, "exp", "", "Card expiration date (MM/YY) (required)")
	cmd.Flags().StringVar(&cvv, "cvv", "", "Card CVV code (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}

// NewAddBinaryCommand creates the "add-binary" command to add binary data secret.
// Flags:
// - client-pub-key-path: Path to client public key
// - token: JWT token for authentication (required)
// - name: Name of the secret (required)
// - data-path: Path to binary data file (required)
// - meta: Optional metadata
func NewClientAddBinaryCommand() *cobra.Command {
	var (
		clientPubKeyFile string
		token            string
		secretName       string
		dataPath         string
		meta             string
	)

	cmd := &cobra.Command{
		Use:   "add-binary",
		Short: "Add a new binary data entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			app, err := apps.NewBinarySecretAddApp(clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to initialize add app: %w", err)
			}

			data, err := os.ReadFile(dataPath)
			if err != nil {
				return fmt.Errorf("failed to read binary data file: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			secret := &models.BinarySecretAdd{
				SecretName:  secretName,
				SecretType:  models.SecretTypeBinary,
				SecretOwner: token,
				Data:        data,
				Meta:        metaPtr,
			}

			if err := app.Execute(ctx, secret, token); err != nil {
				return fmt.Errorf("failed to save binary secret: %w", err)
			}

			fmt.Println("Binary secret saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&clientPubKeyFile, "client-pub-key-path", "", "Path to client public key")
	cmd.Flags().StringVar(&token, "token", "", "JWT token for authentication (required)")
	cmd.Flags().StringVar(&secretName, "name", "", "Name of the secret (required)")
	cmd.Flags().StringVar(&dataPath, "data-path", "", "Path to binary data file (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}

// NewAddTextCommand creates the "add-text" command to add text secret.
// Flags:
// - client-pub-key-path: Path to client public key
// - token: JWT token for authentication (required)
// - name: Name of the secret (required)
// - text: Text content (required)
// - meta: Optional metadata
func NewClientAddTextCommand() *cobra.Command {
	var (
		clientPubKeyFile string
		token            string
		secretName       string
		text             string
		meta             string
	)

	cmd := &cobra.Command{
		Use:   "add-text",
		Short: "Add a new text entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			app, err := apps.NewTextSecretAddApp(clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to initialize add app: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			secret := &models.TextSecretAdd{
				SecretName:  secretName,
				SecretType:  models.SecretTypeText,
				SecretOwner: token,
				Text:        text,
				Meta:        metaPtr,
			}

			if err := app.Execute(ctx, secret, token); err != nil {
				return fmt.Errorf("failed to save text secret: %w", err)
			}

			fmt.Println("Text secret saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&clientPubKeyFile, "client-pub-key-path", "", "Path to client public key")
	cmd.Flags().StringVar(&token, "token", "", "JWT token for authentication (required)")
	cmd.Flags().StringVar(&secretName, "name", "", "Name of the secret (required)")
	cmd.Flags().StringVar(&text, "text", "", "Text content (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}

// NewAddUserCommand creates the "add-user" command to add a user secret.
// Flags:
// - client-pub-key-path: Path to client public key
// - token: JWT token for authentication (required)
// - name: Name of the secret (required)
// - username: Username (required)
// - password: Password (required)
// - meta: Optional metadata
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
		Short: "Add a new user secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			app, err := apps.NewUserSecretAddApp(clientPubKeyFile)
			if err != nil {
				return fmt.Errorf("failed to initialize add app: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			secret := &models.UserSecretAdd{
				SecretName:  secretName,
				SecretType:  models.SecretTypeUser,
				SecretOwner: token,
				Username:    username,
				Password:    password,
				Meta:        metaPtr,
			}

			if err := app.Execute(ctx, secret, token); err != nil {
				return fmt.Errorf("failed to save user secret: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&clientPubKeyFile, "client-pub-key-path", "", "Path to client public key")
	cmd.Flags().StringVar(&token, "token", "", "JWT token for authentication (required)")
	cmd.Flags().StringVar(&secretName, "name", "", "Name of the secret (required)")
	cmd.Flags().StringVar(&username, "username", "", "Username (required)")
	cmd.Flags().StringVar(&password, "password", "", "Password (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}

// NewListCommand creates the "list" command to list stored secrets from the server.
// Flags:
// - server-url: URL of the server (required)
// - token: Authentication token (required)
// - client-pub-key-path: Path to client's public key
// - client-priv-key-path: Path to client's private key
func NewClientListCommand() *cobra.Command {
	var serverURL string
	var privKeyPath string
	var token string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stored secrets from server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			app, err := apps.NewClientListApp(serverURL, privKeyPath)
			if err != nil {
				return fmt.Errorf("failed to initialize client list app: %w", err)
			}
			defer app.Close()

			req := &models.SecretListRequest{Token: token}

			result, err := app.Execute(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to list secrets: %w", err)
			}

			cmd.Println(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL of the server (required)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token (required)")
	cmd.Flags().StringVar(&privKeyPath, "client-priv-key-path", "", "Path to client private key")

	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("token")

	return cmd
}

// NewSyncCommand creates the "sync" command that syncs local secrets with the server.
// Flags:
// - server-url: URL of the server (required)
// - token: Authentication token (required)
// - mode: Sync mode; one of "client", "server", or "interactive"
func NewClientSyncCommand() *cobra.Command {
	var serverURL string
	var token string
	var mode string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync local data with the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var (
				err error
			)

			switch mode {
			case models.SyncModeClient:
				app, err := apps.NewClientSyncClientApp(serverURL)
				if err != nil {
					return fmt.Errorf("failed to initialize client sync app: %w", err)
				}
				defer app.Close()

				err = app.Execute(ctx, token)
				if err != nil {
					return err
				}

			case models.SyncModeServer:
				app := apps.NewClientSyncServerApp()
				// No Close method on Server app, so no defer Close here
				err = app.Execute(ctx, token)
				if err != nil {
					return err
				}

			case models.SyncModeInteractive:
				app, err := apps.NewClientSyncInteractiveApp(serverURL)
				if err != nil {
					return fmt.Errorf("failed to initialize interactive sync app: %w", err)
				}
				defer app.Close()

				err = app.Execute(ctx, cmd.InOrStdin(), token)
				if err != nil {
					return err
				}

			default:
				return fmt.Errorf("invalid mode: %s", mode)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL of the server (required)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token (required)")
	cmd.Flags().StringVar(&mode, "mode", "client", "Sync mode: client, server, interactive")

	cmd.MarkFlagRequired("server-url")
	cmd.MarkFlagRequired("token")

	return cmd
}
