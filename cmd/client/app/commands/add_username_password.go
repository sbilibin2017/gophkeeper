package commands

import (
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/validation"
	"github.com/spf13/cobra"
)

// RegisterAddUsernamePasswordCommand registers the 'add-username-password' command.
func RegisterAddUsernamePasswordCommand(root *cobra.Command) {
	var (
		secretName    string
		user          string
		pass          string
		meta          string
		authURL       string
		tlsClientCert string
		tlsClientKey  string
	)

	cmd := &cobra.Command{
		Use:   "add-username-password",
		Short: "Add a username-password secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validation.ValidateSecretName(secretName); err != nil {
				return fmt.Errorf("invalid secret name: %w", err)
			}
			if err := validation.ValidateMeta(meta); err != nil {
				return fmt.Errorf("invalid meta: %w", err)
			}

			cfg, err := config.NewClientConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

			req := models.UsernamePasswordAddRequest{
				SecretName: secretName,
				Username:   user,
				Password:   pass,
			}
			if meta != "" {
				req.Meta = &meta
			}

			ctx := cmd.Context()

			if cfg.DB != nil {
				if err := client.AddUsernamePasswordLocal(ctx, cfg.DB, req); err != nil {
					return err
				}
				cmd.Println("Username-password secret added locally")
				return nil
			}

			return errors.New("no local DB configured for adding username-password secret")
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&user, "user", "", "Username for username-password secret")
	cmd.Flags().StringVar(&pass, "pass", "", "Password for username-password secret")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("pass")

	root.AddCommand(cmd)
}
