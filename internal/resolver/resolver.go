package resolver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// EncryptedSecretReader defines read operations for EncryptedSecrets.
type ClientSecretReader interface {
	List(ctx context.Context) ([]*models.EncryptedSecret, error)
}

// EncryptedSecretReader defines read operations for EncryptedSecrets.
type ServerSecretReader interface {
	Get(ctx context.Context, secretName string) (*models.EncryptedSecret, error)
}

// EncryptedSecretWriter defines write operations for EncryptedSecrets.
type ServerSecretWriter interface {
	Save(ctx context.Context, secret *models.EncryptedSecret) error
}

type Cryptor interface {
	Decrypt(enc *cryptor.Encrypted) ([]byte, error)
}

// Resolver handles sync operations for secrets.
type Resolver struct {
	clientReader ClientSecretReader
	serverReader ServerSecretReader
	serverWriter ServerSecretWriter
	cryptor      Cryptor
}

// NewResolver creates a new Resolver instance.
func NewResolver(
	clientReader ClientSecretReader,
	serverReader ServerSecretReader,
	serverWriter ServerSecretWriter,
	cryptor Cryptor,
) *Resolver {
	return &Resolver{
		clientReader: clientReader,
		serverReader: serverReader,
		serverWriter: serverWriter,
		cryptor:      cryptor,
	}
}

// ResolveClient syncs client secrets to the server if client version is newer or server version does not exist.
func (r *Resolver) ResolveClient(ctx context.Context) error {
	clientSecrets, err := r.clientReader.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list client secrets: %w", err)
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := r.serverReader.Get(ctx, clientSecret.SecretName)
		if err != nil {
			return err
		}

		if serverSecret == nil || serverSecret.Timestamp < clientSecret.Timestamp {
			if err := r.serverWriter.Save(ctx, clientSecret); err != nil {
				return fmt.Errorf("failed to save client secret '%s' to server: %w", clientSecret.SecretName, err)
			}
		}
	}

	return nil
}

// ResolveServer is a placeholder; no syncing from server to client here.
func (r *Resolver) ResolveServer(ctx context.Context) error {
	return nil
}

// ResolveInteractive lets user decide conflicts interactively.
func (r *Resolver) ResolveInteractive(ctx context.Context, reader io.Reader) error {
	scanner := bufio.NewScanner(reader)

	clientSecrets, err := r.clientReader.List(ctx)
	if err != nil {
		return err
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := r.serverReader.Get(ctx, clientSecret.SecretName)
		if err != nil {
			return err
		}

		if serverSecret == nil {
			fmt.Printf("Conflict for [%s]:\n", clientSecret.SecretName)
			fmt.Println("Server version: <not found>")

			clientPlain, err := r.cryptor.Decrypt(&cryptor.Encrypted{
				Ciphertext: clientSecret.Ciphertext,
				AESKeyEnc:  clientSecret.AESKeyEnc,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Client version (updated at %v):\n%s\n\n", clientSecret.Timestamp, string(clientPlain))

			fmt.Println("No server secret to save, saving client version automatically.")

			if err := r.serverWriter.Save(ctx, clientSecret); err != nil {
				return fmt.Errorf("failed to save client secret '%s': %w", clientSecret.SecretName, err)
			}

			continue // skip to next secret
		}

		// If serverSecret exists, compare timestamps and interactively resolve conflicts only if client is newer or same timestamp
		if clientSecret.Timestamp >= serverSecret.Timestamp {
			fmt.Printf("Conflict for [%s]:\n", clientSecret.SecretName)

			clientPlain, err := r.cryptor.Decrypt(&cryptor.Encrypted{
				Ciphertext: clientSecret.Ciphertext,
				AESKeyEnc:  clientSecret.AESKeyEnc,
			})
			if err != nil {
				return err
			}
			clientText := string(clientPlain)

			serverPlain, err := r.cryptor.Decrypt(&cryptor.Encrypted{
				Ciphertext: serverSecret.Ciphertext,
				AESKeyEnc:  serverSecret.AESKeyEnc,
			})
			if err != nil {
				return err
			}
			serverText := string(serverPlain)

			fmt.Printf("1) Client version (updated at %v):\n%s\n\n", clientSecret.Timestamp, clientText)
			fmt.Printf("2) Server version (updated at %v):\n%s\n\n", serverSecret.Timestamp, serverText)

			fmt.Print("Choose version to keep (1 or 2): ")

			if !scanner.Scan() {
				return errors.New("failed to read input")
			}

			choice := strings.TrimSpace(scanner.Text())

			switch choice {
			case "1":
				if err := r.serverWriter.Save(ctx, clientSecret); err != nil {
					return fmt.Errorf("failed to save client secret '%s': %w", clientSecret.SecretName, err)
				}
			case "2":
				// keep server secret; do nothing
			default:
				return errors.New("invalid choice")
			}
		}
	}

	return nil
}
