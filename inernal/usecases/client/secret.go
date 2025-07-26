package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/inernal/configs/cryptor"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
)

// SecretClientReader defines read operations for encrypted secrets on the local client.
type SecretClientReader interface {
	List(ctx context.Context, filter *models.SecretListFilterDB) ([]*models.SecretDB, error)
}

// SecretClientWriter defines methods for saving encrypted secrets on the client.
type SecretClientWriter interface {
	Save(ctx context.Context, secret *models.SecretDB) error
}

// SecretServerReader defines methods for retrieving secrets from the server.
type SecretServerReader interface {
	Get(ctx context.Context, req *models.SecretGetRequest) (*models.SecretDB, error)
	List(ctx context.Context, req *models.SecretListRequest) ([]*models.SecretDB, error)
}

// SecretServerWriter defines methods for saving secrets to the remote server.
type SecretServerWriter interface {
	Save(ctx context.Context, secret *models.SecretSaveRequest) error
}

// Encryptor defines decryption operations for encrypted secrets.
type Encryptor interface {
	Encrypt(plaintext []byte) (*cryptor.Encrypted, error)
}

// Decryptor defines decryption operations for encrypted secrets.
type Decryptor interface {
	Decrypt(enc *cryptor.Encrypted) ([]byte, error)
}

// SecretClientAddUsecase handles logic for transforming and saving secrets on the client side.
type SecretClientAddUsecase struct {
	writer    SecretClientWriter
	encryptor Encryptor
}

func (uc *SecretClientAddUsecase) AddBankCard(
	ctx context.Context,
	secret *models.BankcardSecretAdd,
	token string,
) error {
	secretJSON, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	enc, err := uc.encryptor.Encrypt(secretJSON)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	now := time.Now()

	secretDB := &models.SecretDB{
		SecretName:  secret.SecretName,
		SecretType:  models.SecretTypeBankCard,
		SecretOwner: token,
		Ciphertext:  enc.Ciphertext,
		AESKeyEnc:   enc.AESKeyEnc,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return uc.writer.Save(ctx, secretDB)
}

// AddUserSecret encrypts and saves a UserSecretAdd.
func (uc *SecretClientAddUsecase) AddUserSecret(
	ctx context.Context,
	secret *models.UserSecretAdd,
	token string,
) error {
	secretJSON, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	enc, err := uc.encryptor.Encrypt(secretJSON)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	now := time.Now()

	secretDB := &models.SecretDB{
		SecretName:  secret.SecretName,
		SecretType:  models.SecretTypeUser,
		SecretOwner: token,
		Ciphertext:  enc.Ciphertext,
		AESKeyEnc:   enc.AESKeyEnc,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return uc.writer.Save(ctx, secretDB)
}

// AddBinarySecret encrypts and saves a BinarySecretAdd.
func (uc *SecretClientAddUsecase) AddBinarySecret(
	ctx context.Context,
	secret *models.BinarySecretAdd,
	token string,
) error {
	secretJSON, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	enc, err := uc.encryptor.Encrypt(secretJSON)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	now := time.Now()

	secretDB := &models.SecretDB{
		SecretName:  secret.SecretName,
		SecretType:  models.SecretTypeBinary,
		SecretOwner: token,
		Ciphertext:  enc.Ciphertext,
		AESKeyEnc:   enc.AESKeyEnc,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return uc.writer.Save(ctx, secretDB)
}

// AddTextSecret encrypts and saves a TextSecretAdd.
func (uc *SecretClientAddUsecase) AddTextSecret(
	ctx context.Context,
	secret *models.TextSecretAdd,
	token string,
) error {
	secretJSON, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	enc, err := uc.encryptor.Encrypt(secretJSON)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	now := time.Now()

	secretDB := &models.SecretDB{
		SecretName:  secret.SecretName,
		SecretType:  models.SecretTypeText,
		SecretOwner: token,
		Ciphertext:  enc.Ciphertext,
		AESKeyEnc:   enc.AESKeyEnc,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return uc.writer.Save(ctx, secretDB)
}

// SecretClientReadUsecase handles logic for retrieving secrets from the server.
type SecretClientReadUsecase struct {
	reader    SecretServerReader
	decryptor Decryptor
}

// NewSecretClientReadUsecase returns a new instance of SecretClientReadUsecase.
func NewSecretClientReadUsecase(
	reader SecretServerReader,
	decryptor Decryptor,
) *SecretClientReadUsecase {
	return &SecretClientReadUsecase{
		reader:    reader,
		decryptor: decryptor,
	}
}

func (uc *SecretClientReadUsecase) List(
	ctx context.Context,
	req *models.SecretListRequest,
) (string, error) {
	encryptedSecrets, err := uc.reader.List(ctx, req)
	if err != nil {
		return "", err
	}

	var decryptedPayloads []interface{}

	for _, encSecret := range encryptedSecrets {
		plaintext, err := uc.decryptor.Decrypt(&cryptor.Encrypted{
			Ciphertext: encSecret.Ciphertext,
			AESKeyEnc:  encSecret.AESKeyEnc,
		})
		if err != nil {
			return "", fmt.Errorf("failed to decrypt secret '%s': %w", encSecret.SecretName, err)
		}

		switch encSecret.SecretType {
		case models.SecretTypeBankCard:
			var payload models.BankcardSecretPayload
			if err := json.Unmarshal(plaintext, &payload); err != nil {
				return "", fmt.Errorf("failed to unmarshal BankcardSecretPayload for '%s': %w", encSecret.SecretName, err)
			}
			decryptedPayloads = append(decryptedPayloads, &payload)

		case models.SecretTypeUser:
			var payload models.UserSecretAdd
			if err := json.Unmarshal(plaintext, &payload); err != nil {
				return "", fmt.Errorf("failed to unmarshal UserSecretAdd for '%s': %w", encSecret.SecretName, err)
			}
			decryptedPayloads = append(decryptedPayloads, &payload)

		case models.SecretTypeBinary:
			var payload models.BinarySecretAdd
			if err := json.Unmarshal(plaintext, &payload); err != nil {
				return "", fmt.Errorf("failed to unmarshal BinarySecretAdd for '%s': %w", encSecret.SecretName, err)
			}
			decryptedPayloads = append(decryptedPayloads, &payload)

		case models.SecretTypeText:
			var payload models.TextSecretAdd
			if err := json.Unmarshal(plaintext, &payload); err != nil {
				return "", fmt.Errorf("failed to unmarshal TextSecretAdd for '%s': %w", encSecret.SecretName, err)
			}
			decryptedPayloads = append(decryptedPayloads, &payload)

		default:
			return "", fmt.Errorf("unsupported secret type: %s", encSecret.SecretType)
		}
	}

	resultBytes, err := json.MarshalIndent(decryptedPayloads, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal decrypted secrets: %w", err)
	}

	return string(resultBytes), nil
}

// SecretSyncUsecase handles synchronization between local client secrets and remote server secrets.
type SecretSyncUsecase struct {
	clientReader SecretClientReader
	serverReader SecretServerReader
	serverWriter SecretServerWriter
	decryptor    Decryptor
}

// NewSecretSyncUsecase constructs a new SecretSyncUsecase with the given dependencies.
func NewSecretSyncUsecase(
	clientReader SecretClientReader,
	serverReader SecretServerReader,
	serverWriter SecretServerWriter,
	decryptor Decryptor,
) *SecretSyncUsecase {
	return &SecretSyncUsecase{
		clientReader: clientReader,
		serverReader: serverReader,
		serverWriter: serverWriter,
		decryptor:    decryptor,
	}
}

// ResolveClient pushes newer local secrets to the server.
func (uc *SecretSyncUsecase) ResolveClient(ctx context.Context, token string) error {
	clientSecrets, err := uc.clientReader.List(ctx, &models.SecretListFilterDB{SecretOwner: token})
	if err != nil {
		return fmt.Errorf("failed to list client secrets: %w", err)
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := uc.serverReader.Get(ctx, &models.SecretGetRequest{
			SecretName: clientSecret.SecretName,
			SecretType: clientSecret.SecretType,
			Token:      token,
		})
		if err != nil {
			return err
		}

		if serverSecret == nil || serverSecret.UpdatedAt.Before(clientSecret.UpdatedAt) {
			if err := uc.serverWriter.Save(ctx, &models.SecretSaveRequest{
				SecretName: clientSecret.SecretName,
				SecretType: clientSecret.SecretType,
				Ciphertext: clientSecret.Ciphertext,
				AESKeyEnc:  clientSecret.AESKeyEnc,
				Token:      token,
			}); err != nil {
				return fmt.Errorf("failed to save client secret '%s' to server: %w", clientSecret.SecretName, err)
			}
		}
	}

	return nil
}

// ResolveServer currently does nothing.
func (uc *SecretSyncUsecase) ResolveServer(ctx context.Context, token string) error {
	// No-op for server sync
	return nil
}

// ResolveInteractive prompts user to resolve conflicts interactively.
func (uc *SecretSyncUsecase) ResolveInteractive(
	ctx context.Context,
	reader io.Reader,
	token string,
) error {
	scanner := bufio.NewScanner(reader)

	clientSecrets, err := uc.clientReader.List(ctx, &models.SecretListFilterDB{SecretOwner: token})
	if err != nil {
		return fmt.Errorf("failed to list client secrets: %w", err)
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := uc.serverReader.Get(ctx, &models.SecretGetRequest{
			SecretName: clientSecret.SecretName,
			SecretType: clientSecret.SecretType,
			Token:      token,
		})
		if err != nil {
			return err
		}

		if serverSecret == nil {
			fmt.Printf("Conflict for [%s]:\n", clientSecret.SecretName)
			fmt.Println("Server version: <not found>")

			clientPlain, err := uc.decryptor.Decrypt(&cryptor.Encrypted{
				Ciphertext: clientSecret.Ciphertext,
				AESKeyEnc:  clientSecret.AESKeyEnc,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Client version (updated at %v):\n%s\n\n", clientSecret.UpdatedAt, string(clientPlain))

			fmt.Println("No server secret to save, saving client version automatically.")

			if err := uc.serverWriter.Save(ctx, &models.SecretSaveRequest{
				SecretName: clientSecret.SecretName,
				SecretType: clientSecret.SecretType,
				Ciphertext: clientSecret.Ciphertext,
				AESKeyEnc:  clientSecret.AESKeyEnc,
				Token:      token,
			}); err != nil {
				return fmt.Errorf("failed to save client secret '%s': %w", clientSecret.SecretName, err)
			}

			continue
		}

		if !clientSecret.UpdatedAt.Before(serverSecret.UpdatedAt) {
			fmt.Printf("Conflict for [%s]:\n", clientSecret.SecretName)

			clientPlain, err := uc.decryptor.Decrypt(&cryptor.Encrypted{
				Ciphertext: clientSecret.Ciphertext,
				AESKeyEnc:  clientSecret.AESKeyEnc,
			})
			if err != nil {
				return err
			}
			clientText := string(clientPlain)

			serverPlain, err := uc.decryptor.Decrypt(&cryptor.Encrypted{
				Ciphertext: serverSecret.Ciphertext,
				AESKeyEnc:  serverSecret.AESKeyEnc,
			})
			if err != nil {
				return err
			}
			serverText := string(serverPlain)

			fmt.Printf("1) Client version (updated at %v):\n%s\n\n", clientSecret.UpdatedAt, clientText)
			fmt.Printf("2) Server version (updated at %v):\n%s\n\n", serverSecret.UpdatedAt, serverText)

			fmt.Print("Choose version to keep (1 or 2): ")

			if !scanner.Scan() {
				return errors.New("failed to read input")
			}

			choice := strings.TrimSpace(scanner.Text())

			switch choice {
			case "1":
				if err := uc.serverWriter.Save(ctx, &models.SecretSaveRequest{
					SecretName: clientSecret.SecretName,
					SecretType: clientSecret.SecretType,
					Ciphertext: clientSecret.Ciphertext,
					AESKeyEnc:  clientSecret.AESKeyEnc,
					Token:      token,
				}); err != nil {
					return fmt.Errorf("failed to save client secret '%s': %w", clientSecret.SecretName, err)
				}
			case "2":
				// Keep server version, do nothing
			default:
				return errors.New("invalid choice")
			}
		}
	}

	return nil
}
