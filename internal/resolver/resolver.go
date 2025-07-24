package resolver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// ClientSecretReader returns a list of client-side secrets as []*models.EncryptedSecret
type ClientSecretReader interface {
	List(ctx context.Context) ([]*models.EncryptedSecret, error)
}

// ServerSecretReader retrieves a single server-side secret by name
type ServerSecretReader interface {
	Get(ctx context.Context, secretName string) (*models.EncryptedSecret, error)
}

// ServerSecretWriter persists a secret (client overwrites server or vice versa)
type ServerSecretWriter interface {
	Save(ctx context.Context, secret *models.EncryptedSecret) error
}

// Resolver handles sync operations for secrets.
type Resolver struct {
	clientReader ClientSecretReader
	serverReader ServerSecretReader
	serverWriter ServerSecretWriter
}

// NewResolver creates a new Resolver instance.
func NewResolver(
	clientReader ClientSecretReader,
	serverReader ServerSecretReader,
	serverWriter ServerSecretWriter,
) *Resolver {
	return &Resolver{
		clientReader: clientReader,
		serverReader: serverReader,
		serverWriter: serverWriter,
	}
}

// ResolveClient syncs client secrets to the server if client version is newer or server version does not exist.
func (r *Resolver) ResolveClient(ctx context.Context) error {
	clientSecrets, err := r.clientReader.List(ctx)
	if err != nil {
		return err
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := r.serverReader.Get(ctx, clientSecret.SecretName)
		if err != nil {
			return err
		}

		if serverSecret == nil || serverSecret.Timestamp < clientSecret.Timestamp {
			if err := r.serverWriter.Save(ctx, clientSecret); err != nil {
				return fmt.Errorf("failed to save client secret '%s': %w", clientSecret.SecretName, err)
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

		if serverSecret == nil || clientSecret.Timestamp >= serverSecret.Timestamp {
			fmt.Printf("Conflict for [%s]:\n", clientSecret.SecretName)

			if serverSecret == nil {
				fmt.Println("Server version: <not found>")
			} else {
				fmt.Printf("1) Client version updated at %v\n", clientSecret.Timestamp)
				fmt.Printf("2) Server version updated at %v\n", serverSecret.Timestamp)
			}

			fmt.Print("Choose version to keep (1 or 2): ")

			if !scanner.Scan() {
				return errors.New("failed to read input")
			}

			choice := strings.TrimSpace(scanner.Text())

			switch choice {
			case "1":
				if err := r.serverWriter.Save(ctx, clientSecret); err != nil {
					return err
				}
			case "2":
			default:
				return errors.New("invalid choice")
			}
		}
	}

	return nil
}
