package commands

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/validation"
)

// RegisterAddBankCardCommand registers the 'add-bank-card' command.
func RegisterAddBankCardCommand(root *cobra.Command) {
	var (
		secretName    string
		number        string
		owner         string
		exp           string
		cvv           string
		meta          string
		authURL       string
		tlsClientCert string
		tlsClientKey  string
	)

	cmd := &cobra.Command{
		Use:   "add-bank-card",
		Short: "Add a bank card secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validation.ValidateSecretName(secretName); err != nil {
				return err
			}
			if err := validation.ValidateBankCardNumber(number); err != nil {
				return err
			}
			if err := validation.ValidateBankCardOwner(owner); err != nil {
				return err
			}
			if err := validation.ValidateBankCardExp(exp); err != nil {
				return err
			}
			if err := validation.ValidateBankCardCVV(cvv); err != nil {
				return err
			}
			if err := validation.ValidateMeta(meta); err != nil {
				return err
			}

			cfg, err := config.NewClientConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return err
			}

			req := models.BankCardAddRequest{
				SecretName: secretName,
				Number:     number,
				Owner:      owner,
				Exp:        exp,
				CVV:        cvv,
			}
			if meta != "" {
				req.Meta = &meta
			}

			ctx := cmd.Context()

			if cfg.DB != nil {
				if err := client.AddBankCardLocal(ctx, cfg.DB, req); err != nil {
					return err
				}
				cmd.Println("Bank card secret added locally")
				return nil
			}

			return errors.New("no local DB configured for adding bank card")
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&number, "number", "", "Bank card number")
	cmd.Flags().StringVar(&owner, "owner", "", "Bank card owner")
	cmd.Flags().StringVar(&exp, "exp", "", "Bank card expiration date")
	cmd.Flags().StringVar(&cvv, "cvv", "", "Bank card CVV code")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file")

	root.AddCommand(cmd)
}
