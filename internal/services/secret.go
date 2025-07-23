package services

import (
	"context"
	"encoding/json"

	"github.com/sbilibin2017/gophkeeper/internal/encryption"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

type Lister interface {
	List(ctx context.Context) ([]*models.EncryptedSecret, error)
}

type Decryptor interface {
	Decrypt(encrypted *encryption.Encrypted) ([]byte, error)
}

type SecretClientReadService struct {
	lister    Lister
	decryptor Decryptor
}

func (s *SecretClientReadService) List(ctx context.Context) (
	[]*models.BankCard,
	[]*models.Binary,
	[]*models.Text,
	[]*models.User,
	error,
) {
	encryptedSecrets, err := s.lister.List(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var (
		bankCards []*models.BankCard
		binaries  []*models.Binary
		texts     []*models.Text
		users     []*models.User
	)

	for _, es := range encryptedSecrets {
		enc := &encryption.Encrypted{
			Ciphertext: es.Ciphertext,
			AESKeyEnc:  es.AESKeyEnc,
		}

		plaintext, err := s.decryptor.Decrypt(enc)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		switch es.SecretType {
		case models.SecretTypeBankCard:
			var card models.BankCard
			if err := json.Unmarshal(plaintext, &card); err != nil {
				return nil, nil, nil, nil, err
			}
			bankCards = append(bankCards, &card)

		case models.SecretTypeBinary:
			var bin models.Binary
			if err := json.Unmarshal(plaintext, &bin); err != nil {
				return nil, nil, nil, nil, err
			}
			binaries = append(binaries, &bin)

		case models.SecretTypeText:
			var txt models.Text
			if err := json.Unmarshal(plaintext, &txt); err != nil {
				return nil, nil, nil, nil, err
			}
			texts = append(texts, &txt)

		case models.SecretTypeUser:
			var usr models.User
			if err := json.Unmarshal(plaintext, &usr); err != nil {
				return nil, nil, nil, nil, err
			}
			users = append(users, &usr)

		default:
			continue
		}
	}

	return bankCards, binaries, texts, users, nil
}
