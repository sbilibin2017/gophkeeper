package bankcard

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/client/config"
	"github.com/sbilibin2017/gophkeeper/internal/client/models"
	"github.com/sbilibin2017/gophkeeper/internal/client/repositories/bankcard"
)

// AddClient saves or updates a bank card locally using the default client DB config.
func AddClient(
	ctx context.Context,
	secretName, number, owner, exp, cvv, meta string,
) error {
	cfg, err := config.NewConfig(config.WithDB())
	if err != nil {
		return err
	}
	defer cfg.DB.Close()

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

	repo := bankcard.NewSaveRepository(cfg.DB)
	if err := repo.Save(ctx, req); err != nil {
		return err
	}

	return nil
}
