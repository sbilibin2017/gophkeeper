package bankcard

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/bankcard"

	"github.com/spf13/cobra"
)

// RegisterAddBankCardCommand adds the 'add-bankcard' command to the root Cobra command.
func RegisterAddBankCardCommand(
	root *cobra.Command,
	runFunc func(ctx context.Context, secretName, number, owner, exp, cvv, meta string) error,
) {
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
			return runFunc(
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

// RunAddBankCard returns a closure that saves a bank card secret to the local encrypted DB.
// It initializes DB connection using db.NewDB and creates repository before returning the actual run function.
func RunAddBankCard(ctx context.Context, secretName, number, owner, exp, cvv, meta string) error {
	dbConn, err := db.NewDB("sqlite", "client.db")
	if err != nil {
		return err
	}
	defer dbConn.Close()

	req := &models.BankCardAddRequest{
		SecretName: secretName,
		Number:     number,
		Owner:      owner,
		Exp:        exp,
		CVV:        cvv,
	}
	if meta != "" {
		req.Meta = &meta
	}

	repo := bankcard.NewSaveRepository(dbConn)
	return repo.Save(ctx, req)
}
