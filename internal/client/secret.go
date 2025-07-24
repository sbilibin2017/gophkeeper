// Package client provides high-level interfaces and implementations for securely
// reading and writing encrypted secrets.
package client

import (
	"context"
	"encoding/json"

	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Reader defines the interface to retrieve encrypted secrets.
type Reader interface {
	Get(ctx context.Context, secretName string) (*models.EncryptedSecret, error)
	List(ctx context.Context) ([]*models.EncryptedSecret, error)
}

// Decryptor defines the interface to decrypt data.
type Decryptor interface {
	Decrypt(encrypted *cryptor.Encrypted) ([]byte, error)
}

// SecretReader provides methods to fetch and decrypt secrets.
type SecretReader struct {
	reader    Reader
	decryptor Decryptor
}

// NewSecretReader constructs a new SecretReader.
func NewSecretReader(reader Reader, decryptor Decryptor) *SecretReader {
	return &SecretReader{
		reader:    reader,
		decryptor: decryptor,
	}
}

// Get retrieves and decrypts a single secret by name.
func (s *SecretReader) Get(ctx context.Context, secretName string) (*string, error) {
	encryptedSecret, err := s.reader.Get(ctx, secretName)
	if err != nil {
		return nil, err
	}
	if encryptedSecret == nil {
		return nil, nil
	}

	enc := &cryptor.Encrypted{
		Ciphertext: encryptedSecret.Ciphertext,
		AESKeyEnc:  encryptedSecret.AESKeyEnc,
	}

	plaintext, err := s.decryptor.Decrypt(enc)
	if err != nil {
		return nil, err
	}

	var indentedJSON []byte

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

// List retrieves and decrypts all available secrets.
func (s *SecretReader) List(ctx context.Context) ([]string, error) {
	encryptedSecrets, err := s.reader.List(ctx)
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
			continue
		}

		if err != nil {
			return nil, err
		}

		secrets = append(secrets, string(indentedJSON))
	}

	return secrets, nil
}

// Writer defines the interface to persist or delete a secret.
type Writer interface {
	Save(ctx context.Context, secret *models.EncryptedSecret) error
	Delete(ctx context.Context, secretName string) error
}

// Encryptor defines the interface to encrypt plaintext data.
type Encryptor interface {
	Encrypt(plaintext []byte) (*cryptor.Encrypted, error)
}

// SecretWriter provides methods to encrypt and store secrets.
type SecretWriter struct {
	writer    Writer
	encryptor Encryptor
}

// NewSecretWriter constructs a new SecretWriter.
func NewSecretWriter(writer Writer, encryptor Encryptor) *SecretWriter {
	return &SecretWriter{
		writer:    writer,
		encryptor: encryptor,
	}
}

// AddBankCard encrypts and stores a bank card secret.
func (s *SecretWriter) AddBankCard(ctx context.Context, secretName string, payload models.BankCardPayload) error {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	enc, err := s.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	secret := &models.EncryptedSecret{
		SecretType: models.SecretTypeBankCard,
		SecretName: secretName,
		Ciphertext: enc.Ciphertext,
		AESKeyEnc:  enc.AESKeyEnc,
	}

	return s.writer.Save(ctx, secret)
}

// AddBinary encrypts and stores a binary secret.
func (s *SecretWriter) AddBinary(ctx context.Context, secretName string, payload models.BinaryPayload) error {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	enc, err := s.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	secret := &models.EncryptedSecret{
		SecretType: models.SecretTypeBinary,
		SecretName: secretName,
		Ciphertext: enc.Ciphertext,
		AESKeyEnc:  enc.AESKeyEnc,
	}

	return s.writer.Save(ctx, secret)
}

// AddText encrypts and stores a text secret.
func (s *SecretWriter) AddText(ctx context.Context, secretName string, payload models.TextPayload) error {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	enc, err := s.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	secret := &models.EncryptedSecret{
		SecretType: models.SecretTypeText,
		SecretName: secretName,
		Ciphertext: enc.Ciphertext,
		AESKeyEnc:  enc.AESKeyEnc,
	}

	return s.writer.Save(ctx, secret)
}

// AddUser encrypts and stores a user credential secret.
func (s *SecretWriter) AddUser(ctx context.Context, secretName string, payload models.UserPayload) error {
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	enc, err := s.encryptor.Encrypt(plaintext)
	if err != nil {
		return err
	}

	secret := &models.EncryptedSecret{
		SecretType: models.SecretTypeUser,
		SecretName: secretName,
		Ciphertext: enc.Ciphertext,
		AESKeyEnc:  enc.AESKeyEnc,
	}

	return s.writer.Save(ctx, secret)
}

// Delete removes a secret by name.
func (s *SecretWriter) Delete(ctx context.Context, secretName string) error {
	return s.writer.Delete(ctx, secretName)
}
