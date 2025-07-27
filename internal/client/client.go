package client

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Registerer defines the interface for registering a new user.
type Registerer interface {
	Register(ctx context.Context, username string, password string) (*string, error)
}

// Loginer defines the interface for logging in a user.
type Loginer interface {
	Login(ctx context.Context, username string, password string) (*string, error)
}

// Encryptor defines the interface for encrypting plaintext data.
type Encryptor interface {
	Encrypt(plaintext []byte) (*models.SecretEncrypted, error)
}

// Decryptor defines the interface for decrypting SecretEncrypted secrets.
type Decryptor interface {
	Decrypt(secret *models.SecretEncrypted) ([]byte, error)
}

// ClientSaver defines the interface for saving secrets from the client.
type ClientSaver interface {
	Save(
		ctx context.Context,
		secretOwner string,
		secretName string,
		secretType string,
		ciphertext []byte,
		aesKeyEnc []byte,
	) error
}

// ClientLister defines the interface for listing secrets from the client.
type ClientLister interface {
	List(ctx context.Context, secretOwner string) ([]*models.Secret, error)
}

// ServerGetter defines the interface for retrieving a secret from the server.
type ServerGetter interface {
	Get(
		ctx context.Context,
		secretOwner string,
		secretType string,
		secretName string,
	) (*models.Secret, error)
}

// ServerLister defines the interface for listing secrets from the server.
type ServerLister interface {
	List(ctx context.Context, secretOwner string) ([]*models.Secret, error)
}

// ServerSaver defines the interface for saving secrets to the server.
type ServerSaver interface {
	Save(
		ctx context.Context,
		secretOwner string,
		secretName string,
		secretType string,
		ciphertext []byte,
		aesKeyEnc []byte,
	) error
}

// ClientResolver defines the interface for client-side synchronization of secrets.
type ClientResolver interface {
	Resolve(ctx context.Context, secretOwner string) error
}

// InteractiveResolver defines the interface for interactive resolution of conflicts during sync.
type InteractiveResolver interface {
	Resolve(ctx context.Context, secretOwner string, reader io.Reader) error
}

// ClientRegister registers a new user with a username and password.
// It returns an authentication token on success.
func ClientRegister(
	ctx context.Context,
	registerer Registerer,
	username string,
	password string,
) (string, error) {
	tokenPtr, err := registerer.Register(ctx, username, password)
	if err != nil {
		return "", err
	}
	if tokenPtr == nil {
		return "", errors.New("registration returned nil token")
	}
	return *tokenPtr, nil
}

// ClientLogin logs in an existing user with username and password.
// It returns an authentication token on success.
func ClientLogin(
	ctx context.Context,
	loginer Loginer,
	username string,
	password string,
) (string, error) {
	tokenPtr, err := loginer.Login(ctx, username, password)
	if err != nil {
		return "", err
	}
	if tokenPtr == nil {
		return "", errors.New("login returned nil token")
	}
	return *tokenPtr, nil
}

// ClientAddBankcard encrypts and saves a bankcard secret.
func ClientAddBankcard(
	ctx context.Context,
	clientSaver ClientSaver,
	encryptor Encryptor,
	token string,
	secretName string,
	number string,
	owner string,
	exp string,
	cvv string,
	meta string,
) error {
	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}

	payload := models.BankcardPayload{
		Number: number,
		Owner:  owner,
		Exp:    exp,
		CVV:    cvv,
		Meta:   metaPtr,
	}

	plaintext, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal bankcard payload: %w", err)
	}

	SecretEncrypted, err := encryptor.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return clientSaver.Save(
		ctx,
		token,
		secretName,
		models.SecretTypeBankCard,
		SecretEncrypted.Ciphertext,
		SecretEncrypted.AESKeyEnc,
	)
}

// ClientAddText encrypts and saves a text secret.
func ClientAddText(
	ctx context.Context,
	clientSaver ClientSaver,
	encryptor Encryptor,
	token string,
	secretName string,
	data string,
	meta string,
) error {
	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}

	payload := models.TextPayload{
		Data: data,
		Meta: metaPtr,
	}

	plaintext, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal text payload: %w", err)
	}

	SecretEncrypted, err := encryptor.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return clientSaver.Save(
		ctx,
		token,
		secretName,
		models.SecretTypeText,
		SecretEncrypted.Ciphertext,
		SecretEncrypted.AESKeyEnc,
	)
}

// ClientAddBinary encrypts and saves a binary secret.
// The data is expected to be a base64-encoded string.
func ClientAddBinary(
	ctx context.Context,
	clientSaver ClientSaver,
	encryptor Encryptor,
	token string,
	secretName string,
	data string,
	meta string,
) error {
	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}

	binData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("failed to decode base64 data: %w", err)
	}

	payload := models.BinaryPayload{
		Data: binData,
		Meta: metaPtr,
	}

	plaintext, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal binary payload: %w", err)
	}

	SecretEncrypted, err := encryptor.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return clientSaver.Save(
		ctx,
		token,
		secretName,
		models.SecretTypeBinary,
		SecretEncrypted.Ciphertext,
		SecretEncrypted.AESKeyEnc,
	)
}

// ClientAddUser encrypts and saves a user credential secret.
func ClientAddUser(
	ctx context.Context,
	clientSaver ClientSaver,
	encryptor Encryptor,
	token string,
	secretName string,
	username string,
	password string,
	meta string,
) error {
	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}

	payload := models.UserPayload{
		Username: username,
		Password: password,
		Meta:     metaPtr,
	}

	plaintext, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal user payload: %w", err)
	}

	SecretEncrypted, err := encryptor.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return clientSaver.Save(
		ctx,
		token,
		secretName,
		models.SecretTypeUser,
		SecretEncrypted.Ciphertext,
		SecretEncrypted.AESKeyEnc,
	)
}

// ClientListSecrets fetches, decrypts, and returns secrets associated with the given token.
func ClientListSecrets(
	ctx context.Context,
	secretReader ServerLister,
	decryptor Decryptor,
	token string,
) (string, error) {
	secrets, err := secretReader.List(ctx, token)
	if err != nil {
		return "", err
	}

	var builder strings.Builder

	for _, secret := range secrets {
		decrypted, err := decryptor.Decrypt(&models.SecretEncrypted{
			Ciphertext: secret.Ciphertext,
			AESKeyEnc:  secret.AESKeyEnc,
		})
		if err != nil {
			return "", fmt.Errorf("failed to decrypt secret %s: %w", secret.SecretName, err)
		}

		switch secret.SecretType {
		case models.SecretTypeBankCard:
			var bankcard models.BankcardPayload
			if err := json.Unmarshal(decrypted, &bankcard); err != nil {
				return "", fmt.Errorf("failed to unmarshal bankcard: %w", err)
			}
			out, _ := json.MarshalIndent(bankcard, "", "  ")
			builder.Write(out)

		case models.SecretTypeText:
			var text models.TextPayload
			if err := json.Unmarshal(decrypted, &text); err != nil {
				return "", fmt.Errorf("failed to unmarshal text: %w", err)
			}
			out, _ := json.MarshalIndent(text, "", "  ")
			builder.Write(out)

		case models.SecretTypeBinary:
			var binary models.BinaryPayload
			if err := json.Unmarshal(decrypted, &binary); err != nil {
				return "", fmt.Errorf("failed to unmarshal binary: %w", err)
			}
			out, _ := json.MarshalIndent(binary, "", "  ")
			builder.Write(out)

		case models.SecretTypeUser:
			var user models.UserPayload
			if err := json.Unmarshal(decrypted, &user); err != nil {
				return "", fmt.Errorf("failed to unmarshal user: %w", err)
			}
			out, _ := json.MarshalIndent(user, "", "  ")
			builder.Write(out)

		default:
			builder.WriteString(fmt.Sprintf("Unknown secret type: %s\n", secret.SecretType))
		}
		builder.WriteString("\n\n")
	}

	return builder.String(), nil
}

// Constants as you defined
const (
	ResolveStrategyServer      = "server"
	ResolveStrategyClient      = "client"
	ResolveStrategyInteractive = "interactive"
)

// ClientSyncClient synchronizes secrets with the server using client resolution.
func ClientSyncClient(
	ctx context.Context,
	cl ClientLister,
	sg ServerGetter,
	ss ServerSaver,
	secretOwner string,
) error {
	clientSecrets, err := cl.List(ctx, secretOwner)
	if err != nil {
		return fmt.Errorf("failed to list client secrets: %w", err)
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := sg.Get(ctx, secretOwner, clientSecret.SecretType, clientSecret.SecretName)
		if err != nil {
			return fmt.Errorf("failed to get server secret: %w", err)
		}

		if serverSecret == nil || clientSecret.UpdatedAt.After(serverSecret.UpdatedAt) {
			err := ss.Save(
				ctx,
				secretOwner,
				clientSecret.SecretName,
				clientSecret.SecretType,
				clientSecret.Ciphertext,
				clientSecret.AESKeyEnc,
			)
			if err != nil {
				return fmt.Errorf("failed to save secret to server: %w", err)
			}
		}
	}

	return nil
}

// ClientSyncInteractive synchronizes secrets with the server using interactive resolution via input reader.
func ClientSyncInteractive(
	ctx context.Context,
	cl ClientLister,
	sg ServerGetter,
	ss ServerSaver,
	d Decryptor,
	secretOwner string,
	reader io.Reader,
) error {
	scanner := bufio.NewScanner(reader)

	clientSecrets, err := cl.List(ctx, secretOwner)
	if err != nil {
		return fmt.Errorf("failed to list client secrets: %w", err)
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := sg.Get(ctx, secretOwner, clientSecret.SecretType, clientSecret.SecretName)
		if err != nil {
			return fmt.Errorf("failed to get server secret: %w", err)
		}

		if serverSecret == nil {
			fmt.Printf("Server does not contain secret [%s], uploading client version.\n", clientSecret.SecretName)
			err := ss.Save(
				ctx,
				secretOwner,
				clientSecret.SecretName,
				clientSecret.SecretType,
				clientSecret.Ciphertext,
				clientSecret.AESKeyEnc,
			)
			if err != nil {
				return fmt.Errorf("failed to save client secret: %w", err)
			}
			continue
		}

		if !clientSecret.UpdatedAt.Before(serverSecret.UpdatedAt) {
			clientPlain, err := d.Decrypt(&models.SecretEncrypted{
				Ciphertext: clientSecret.Ciphertext,
				AESKeyEnc:  clientSecret.AESKeyEnc,
			})
			if err != nil {
				log.Printf("failed to decrypt client secret %s: %v\n", clientSecret.SecretName, err)
				continue
			}

			serverPlain, err := d.Decrypt(&models.SecretEncrypted{
				Ciphertext: serverSecret.Ciphertext,
				AESKeyEnc:  serverSecret.AESKeyEnc,
			})
			if err != nil {
				log.Printf("failed to decrypt server secret %s: %v\n", serverSecret.SecretName, err)
				continue
			}

			// Pretty-print client version
			var clientPretty string
			var clientData interface{}
			if err := json.Unmarshal(clientPlain, &clientData); err != nil {
				clientPretty = string(clientPlain)
			} else {
				b, err := json.MarshalIndent(clientData, "", "  ")
				if err != nil {
					clientPretty = string(clientPlain)
				} else {
					clientPretty = string(b)
				}
			}

			// Pretty-print server version
			var serverPretty string
			var serverData interface{}
			if err := json.Unmarshal(serverPlain, &serverData); err != nil {
				serverPretty = string(serverPlain)
			} else {
				b, err := json.MarshalIndent(serverData, "", "  ")
				if err != nil {
					serverPretty = string(serverPlain)
				} else {
					serverPretty = string(b)
				}
			}

			fmt.Printf("Conflict for [%s]:\n", clientSecret.SecretName)
			fmt.Printf("1) Client version (updated at %s):\n%s\n\n", clientSecret.UpdatedAt.Format(time.RFC3339), clientPretty)
			fmt.Printf("2) Server version (updated at %s):\n%s\n\n", serverSecret.UpdatedAt.Format(time.RFC3339), serverPretty)
			fmt.Print("Choose version to keep (1 - client / 2 - server): ")

			if !scanner.Scan() {
				return errors.New("input scan failed")
			}

			switch strings.TrimSpace(scanner.Text()) {
			case "1":
				err := ss.Save(
					ctx,
					secretOwner,
					clientSecret.SecretName,
					clientSecret.SecretType,
					clientSecret.Ciphertext,
					clientSecret.AESKeyEnc,
				)
				if err != nil {
					return fmt.Errorf("failed to save client version: %w", err)
				}
			case "2":
				// Keep server version, do nothing
			default:
				fmt.Println("Invalid input, skipping...")
			}
		}
	}

	return nil
}
