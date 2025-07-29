package usecases

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
)

// ClientLister defines an interface for listing client secrets.
type ClientLister interface {
	// List returns all secrets stored on the client side for the given secret owner.
	List(ctx context.Context, secretOwner string) ([]*models.Secret, error)
}

// ServerGetter defines an interface for retrieving secrets from the server by name.
type ServerGetter interface {
	// Get retrieves a secret by its name from the server.
	Get(
		ctx context.Context,
		token string,
		secretName string,
		secretType string,
	) (*models.Secret, error)
}

// ServerSaver defines an interface for saving secrets to the server.
type ServerSaver interface {
	// Save persists a secret to the server.
	Save(ctx context.Context, token string, secret *models.Secret) error
}

// Cryptor defines an interface for decrypting encrypted secrets.
type Cryptor interface {
	// Decrypt decrypts an encrypted secret returning the plaintext.
	Decrypt(enc *models.SecretEncrypted) ([]byte, error)
}

type ClientSyncUsecase struct {
	clientLister ClientLister
	serverGetter ServerGetter
	serverSaver  ServerSaver
}

func NewClientSyncUsecase(
	clientLister ClientLister,
	serverGetter ServerGetter,
	serverSaver ServerSaver,
) *ClientSyncUsecase {
	return &ClientSyncUsecase{
		clientLister: clientLister,
		serverGetter: serverGetter,
		serverSaver:  serverSaver,
	}
}

func (r *ClientSyncUsecase) Sync(ctx context.Context, token string) error {
	clientSecrets, err := r.clientLister.List(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to list client secrets: %w", err)
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := r.serverGetter.Get(ctx, token, clientSecret.SecretName, clientSecret.SecretType)
		if err != nil {
			return fmt.Errorf("failed to get secret from server: %w", err)
		}

		// If server has no such secret or client's version is newer
		if serverSecret == nil || serverSecret.UpdatedAt.Before(clientSecret.UpdatedAt) {
			if err := r.serverSaver.Save(ctx, token, clientSecret); err != nil {
				return fmt.Errorf("failed to save client secret '%s': %w", clientSecret.SecretName, err)
			}
		}
	}

	return nil
}

// ServerSyncUsecase is a placeholder for server-to-client synchronization logic.
type ServerSyncUsecase struct{}

// NewServerSyncUsecase creates a new ServerSyncUsecase instance.
func NewServerSyncUsecase() *ServerSyncUsecase {
	return &ServerSyncUsecase{}
}

// Sync currently does nothing and returns nil.
func (r *ServerSyncUsecase) Sync(ctx context.Context) error {
	return nil
}

type InteractiveSyncUsecase struct {
	clientLister ClientLister
	serverGetter ServerGetter
	serverSaver  ServerSaver
	cryptor      Cryptor
}

func NewInteractiveSyncUsecase(
	clientLister ClientLister,
	serverGetter ServerGetter,
	serverSaver ServerSaver,
	cryptor Cryptor,
) *InteractiveSyncUsecase {
	return &InteractiveSyncUsecase{
		clientLister: clientLister,
		serverGetter: serverGetter,
		serverSaver:  serverSaver,
		cryptor:      cryptor,
	}
}

func (r *InteractiveSyncUsecase) Sync(ctx context.Context, reader io.Reader, token string) error {
	scanner := bufio.NewScanner(reader)

	clientSecrets, err := r.clientLister.List(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to list client secrets: %w", err)
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := r.serverGetter.Get(ctx, token, clientSecret.SecretName, clientSecret.SecretType)
		if err != nil {
			return fmt.Errorf("failed to get server secret: %w", err)
		}

		if serverSecret == nil {
			fmt.Printf("Secret [%s] not found on server. Saving client version.\n", clientSecret.SecretName)
			if err := r.serverSaver.Save(ctx, token, clientSecret); err != nil {
				return fmt.Errorf("failed to save client secret '%s': %w", clientSecret.SecretName, err)
			}
			continue
		}

		if !serverSecret.UpdatedAt.After(clientSecret.UpdatedAt) {
			clientPlain, err := r.cryptor.Decrypt(&models.SecretEncrypted{
				Ciphertext: clientSecret.Ciphertext,
				AESKeyEnc:  clientSecret.AESKeyEnc,
			})
			if err != nil {
				return fmt.Errorf("decrypt client secret '%s' error: %w", clientSecret.SecretName, err)
			}

			serverPlain, err := r.cryptor.Decrypt(&models.SecretEncrypted{
				Ciphertext: serverSecret.Ciphertext,
				AESKeyEnc:  serverSecret.AESKeyEnc,
			})
			if err != nil {
				return fmt.Errorf("decrypt server secret '%s' error: %w", serverSecret.SecretName, err)
			}

			fmt.Printf("Conflict for [%s]:\n", clientSecret.SecretName)
			fmt.Printf("1) Client (Updated at %v):\n%s\n", clientSecret.UpdatedAt, string(clientPlain))
			fmt.Printf("2) Server (Updated at %v):\n%s\n", serverSecret.UpdatedAt, string(serverPlain))
			fmt.Print("Choose version to keep (1 or 2): ")

			if !scanner.Scan() {
				return errors.New("failed to read input")
			}

			choice := strings.TrimSpace(scanner.Text())
			switch choice {
			case "1":
				if err := r.serverSaver.Save(ctx, token, clientSecret); err != nil {
					return fmt.Errorf("failed to save client version for '%s': %w", clientSecret.SecretName, err)
				}
			case "2":
			default:
				return errors.New("invalid choice")
			}
		}
	}

	return nil
}
