package commands

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/validation"
)

func RegisterAddBinaryCommand(root *cobra.Command) {
	var (
		secretName    string
		dataPath      string
		meta          string
		authURL       string
		tlsClientCert string
		tlsClientKey  string
	)

	cmd := &cobra.Command{
		Use:   "add-binary-secret",
		Short: "Add a binary secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validation.ValidateSecretName(secretName); err != nil {
				return err
			}
			if err := validation.ValidateDataPath(dataPath); err != nil {
				return err
			}
			if err := validation.ValidateMeta(meta); err != nil {
				return err
			}

			cfg, err := config.NewClientConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return err
			}

			data, err := os.ReadFile(dataPath)
			if err != nil {
				return err
			}

			req := models.BinaryAddRequest{
				SecretName: secretName,
				Data:       data,
			}
			if meta != "" {
				req.Meta = &meta
			}

			ctx := cmd.Context()

			if cfg.DB != nil {
				if err := client.AddBinaryLocal(ctx, cfg.DB, req); err != nil {
					return err
				}
				cmd.Println("Binary secret added locally")
				return nil
			}

			return errors.New("no local DB configured for adding binary secret")
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&dataPath, "data-path", "", "Path to binary data file")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file")

	root.AddCommand(cmd)
}
