package bankcard

import (
	"github.com/sbilibin2017/gophkeeper/internal/client/handlers/bankcard"
	"github.com/spf13/cobra"
)

// Injectable for test/mocking
var addBankCardClient = bankcard.AddClient

// RegisterAddBankCardCommand adds the 'add-bankcard' command to the root command.
func RegisterAddCommand(root *cobra.Command) {
	var (
		secretName string
		number     string
		owner      string
		exp        string
		cvv        string
		meta       string
	)

	cmd := &cobra.Command{
		Use:     "add-bankcard",
		Short:   "Add a new bank card locally",
		Long:    "Add a new bank card secret including number, owner, expiry date, CVV, and metadata to the local encrypted database.",
		Example: `gophkeeper add-bankcard --secret-name mycard --number 1234567890123456 --owner "John Doe" --exp 12/25 --cvv 123 --meta "personal card"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return addBankCardClient(
				cmd.Context(),
				secretName,
				number,
				owner,
				exp,
				cvv,
				meta,
			)
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name for bank card")
	cmd.Flags().StringVar(&number, "number", "", "Bank card number")
	cmd.Flags().StringVar(&owner, "owner", "", "Owner of the bank card")
	cmd.Flags().StringVar(&exp, "exp", "", "Expiration date (MM/YY)")
	cmd.Flags().StringVar(&cvv, "cvv", "", "CVV code")
	cmd.Flags().StringVar(&meta, "meta", "", "Additional metadata")

	cmd.MarkFlagRequired("secret-name")
	cmd.MarkFlagRequired("number")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("exp")
	cmd.MarkFlagRequired("cvv")

	root.AddCommand(cmd)
}
