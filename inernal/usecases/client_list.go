package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
)

// ServerLister defines an interface for retrieving a list of secrets.
type ServerLister interface {
	// List returns a list of secrets from the storage based on the request.
	List(ctx context.Context, token string) ([]*models.Secret, error)
}

// Decryptor defines an interface for decrypting encrypted secrets.
type Decryptor interface {
	// Decrypt decrypts the given encrypted data and returns plaintext.
	Decrypt(enc *models.SecretEncrypted) ([]byte, error)
}

// ClientListUsecase is the main application struct that lists and decrypts secrets on the client.
type ClientListUsecase struct {
	lister    ServerLister
	decryptor Decryptor
}

// NewClientListUsecase creates a new instance of ClientListUsecase.
func NewClientListUsecase(
	lister ServerLister,
	decryptor Decryptor,
) *ClientListUsecase {
	return &ClientListUsecase{
		lister:    lister,
		decryptor: decryptor,
	}
}

// Execute lists and decrypts secrets from the client storage, categorizing them by type.
//
// It returns four separate slices for different secret types: bankcards, users, texts, and binaries.
// In case of any error (e.g. decryption or unmarshalling), it returns a non-nil error.
func (a *ClientListUsecase) Execute(
	ctx context.Context,
	token string,
) (
	[]models.Bankcard,
	[]models.User,
	[]models.Text,
	[]models.Binary,
	error,
) {

	secretsList, err := a.lister.List(ctx, token)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var bankcards []models.Bankcard
	var users []models.User
	var texts []models.Text
	var binaries []models.Binary

	for _, secret := range secretsList {
		encrypted := &models.SecretEncrypted{
			Ciphertext: secret.Ciphertext,
			AESKeyEnc:  secret.AESKeyEnc,
		}

		plaintext, derr := a.decryptor.Decrypt(encrypted)
		if derr != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to decrypt secret %s: %w", secret.SecretName, derr)
		}

		switch secret.SecretType {
		case models.SecretTypeBankCard:
			var card models.Bankcard
			if uerr := json.Unmarshal(plaintext, &card); uerr != nil {
				return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal bankcard secret %s: %w", secret.SecretName, uerr)
			}
			bankcards = append(bankcards, card)

		case models.SecretTypeUser:
			var user models.User
			if uerr := json.Unmarshal(plaintext, &user); uerr != nil {
				return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal user secret %s: %w", secret.SecretName, uerr)
			}
			users = append(users, user)

		case models.SecretTypeText:
			var text models.Text
			if uerr := json.Unmarshal(plaintext, &text); uerr != nil {
				return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal text secret %s: %w", secret.SecretName, uerr)
			}
			texts = append(texts, text)

		case models.SecretTypeBinary:
			var binary models.Binary
			if uerr := json.Unmarshal(plaintext, &binary); uerr != nil {
				return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal binary secret %s: %w", secret.SecretName, uerr)
			}
			binaries = append(binaries, binary)

		default:
			return nil, nil, nil, nil, fmt.Errorf("unknown secret type %s for secret %s", secret.SecretType, secret.SecretName)
		}
	}

	return bankcards, users, texts, binaries, nil
}
