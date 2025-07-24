package client

import (
	"context"
	"encoding/json"

	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

type Getter interface {
	Get(ctx context.Context, secretName string) (*models.EncryptedSecret, error)
}

type Lister interface {
	List(ctx context.Context) ([]*models.EncryptedSecret, error)
}

type Decryptor interface {
	Decrypt(encrypted *cryptor.Encrypted) ([]byte, error)
}

type SecretReader struct {
	lister    Lister
	getter    Getter
	decryptor Decryptor
}

func (s *SecretReader) Get(
	ctx context.Context,
	secretName string,
) (*string, error) {
	encryptedSecret, err := s.getter.Get(ctx, secretName)
	if err != nil {
		return nil, err
	}

	enc := &cryptor.Encrypted{
		Ciphertext: encryptedSecret.Ciphertext,
		AESKeyEnc:  encryptedSecret.AESKeyEnc,
	}

	plaintext, err := s.decryptor.Decrypt(enc)
	if err != nil {
		return nil, err
	}

	var (
		indentedJSON []byte
	)

	switch encryptedSecret.SecretType {
	case models.SecretTypeBankCard:
		var card models.BankCard
		if err := json.Unmarshal(plaintext, &card); err != nil {
			return nil, err
		}
		indentedJSON, err = json.MarshalIndent(card, "", "  ")

	case models.SecretTypeBinary:
		var bin models.Binary
		if err := json.Unmarshal(plaintext, &bin); err != nil {
			return nil, err
		}
		indentedJSON, err = json.MarshalIndent(bin, "", "  ")

	case models.SecretTypeText:
		var txt models.Text
		if err := json.Unmarshal(plaintext, &txt); err != nil {
			return nil, err
		}
		indentedJSON, err = json.MarshalIndent(txt, "", "  ")

	case models.SecretTypeUser:
		var usr models.User
		if err := json.Unmarshal(plaintext, &usr); err != nil {
			return nil, err
		}
		indentedJSON, err = json.MarshalIndent(usr, "", "  ")

	default:
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	result := string(indentedJSON)
	return &result, nil
}

func (s *SecretReader) List(ctx context.Context) ([]string, error) {
	encryptedSecrets, err := s.lister.List(ctx)
	if err != nil {
		return nil, err
	}

	var secrets []string

	for _, es := range encryptedSecrets {
		enc := &cryptor.Encrypted{
			Ciphertext: es.Ciphertext,
			AESKeyEnc:  es.AESKeyEnc,
		}

		plaintext, err := s.decryptor.Decrypt(enc)
		if err != nil {
			return nil, err
		}

		var indentedJSON []byte

		switch es.SecretType {
		case models.SecretTypeBankCard:
			var card models.BankCard
			if err := json.Unmarshal(plaintext, &card); err != nil {
				return nil, err
			}
			indentedJSON, err = json.MarshalIndent(card, "", "  ")
			if err != nil {
				return nil, err
			}
		case models.SecretTypeBinary:
			var bin models.Binary
			if err := json.Unmarshal(plaintext, &bin); err != nil {
				return nil, err
			}
			indentedJSON, err = json.MarshalIndent(bin, "", "  ")
			if err != nil {
				return nil, err
			}
		case models.SecretTypeText:
			var txt models.Text
			if err := json.Unmarshal(plaintext, &txt); err != nil {
				return nil, err
			}
			indentedJSON, err = json.MarshalIndent(txt, "", "  ")
			if err != nil {
				return nil, err
			}
		case models.SecretTypeUser:
			var usr models.User
			if err := json.Unmarshal(plaintext, &usr); err != nil {
				return nil, err
			}
			indentedJSON, err = json.MarshalIndent(usr, "", "  ")
			if err != nil {
				return nil, err
			}
		default:
			continue
		}

		secrets = append(secrets, string(indentedJSON))
	}

	return secrets, nil
}

type Saver interface {
	Save(ctx context.Context, secret *models.EncryptedSecret) error
}

type Deleter interface {
	Delete(ctx context.Context, secretName string) error
}

type Encryptor interface {
	Encrypt(plaintext []byte) (*cryptor.Encrypted, error)
}

type SecretWriter struct {
	saver     Saver
	deleter   Deleter
	encryptor Encryptor
}

func (s *SecretWriter) AddBankCard(
	ctx context.Context,
	secretName string,
	payload models.BankCardPayload,
) error {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	enc, err := s.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	encryptedSecret := &models.EncryptedSecret{
		SecretType: models.SecretTypeBankCard,
		SecretName: secretName,
		Ciphertext: enc.Ciphertext,
		AESKeyEnc:  enc.AESKeyEnc,
	}

	return s.saver.Save(ctx, encryptedSecret)
}

func (s *SecretWriter) AddBinary(
	ctx context.Context,
	secretName string,
	payload models.BinaryPayload,
) error {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	enc, err := s.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	encryptedSecret := &models.EncryptedSecret{
		SecretType: models.SecretTypeBinary,
		SecretName: secretName,
		Ciphertext: enc.Ciphertext,
		AESKeyEnc:  enc.AESKeyEnc,
	}

	return s.saver.Save(ctx, encryptedSecret)
}

func (s *SecretWriter) AddText(
	ctx context.Context,
	secretName string,
	payload models.TextPayload,
) error {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	enc, err := s.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	encryptedSecret := &models.EncryptedSecret{
		SecretType: models.SecretTypeText,
		SecretName: secretName,
		Ciphertext: enc.Ciphertext,
		AESKeyEnc:  enc.AESKeyEnc,
	}

	return s.saver.Save(ctx, encryptedSecret)
}

func (s *SecretWriter) AddUser(
	ctx context.Context,
	secretName string,
	payload models.UserPayload,
) error {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	enc, err := s.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	encryptedSecret := &models.EncryptedSecret{
		SecretType: models.SecretTypeUser,
		SecretName: secretName,
		Ciphertext: enc.Ciphertext,
		AESKeyEnc:  enc.AESKeyEnc,
	}

	return s.saver.Save(ctx, encryptedSecret)
}

func (s *SecretWriter) Delete(
	ctx context.Context,
	secretName string,
) error {
	return s.deleter.Delete(ctx, secretName)
}
