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

// Lister returns a list of client-side secrets
type Lister interface {
	List(ctx context.Context) ([]*models.EncryptedSecret, error)
}

// Getter retrieves a single server-side secret by name
type Getter interface {
	Get(ctx context.Context, secretName string) (*models.EncryptedSecret, error)
}

// Saver persists a secret (client overwrites server or vice versa)
type Saver interface {
	Save(ctx context.Context, secret *models.EncryptedSecret) error
}

// Resolver handles sync operations for secrets.
type Resolver struct {
	lister Lister
	getter Getter
	saver  Saver
}

// NewResolver creates a new Resolver instance.
func NewResolver(
	lister Lister,
	getter Getter,
	saver Saver,
) *Resolver {
	return &Resolver{
		lister: lister,
		getter: getter,
		saver:  saver,
	}
}

func (r *Resolver) ResolveClient(ctx context.Context) error {
	clientSecrets, err := r.lister.List(ctx)
	if err != nil {
		return err
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := r.getter.Get(ctx, clientSecret.SecretName)
		if err != nil {
			return err
		}

		if serverSecret == nil || serverSecret.Timestamp < clientSecret.Timestamp {
			if err := r.saver.Save(ctx, clientSecret); err != nil {
				return fmt.Errorf("failed to save client secret: %w", err)
			}
		}
	}
	return nil
}

func (r *Resolver) ResolveServer(ctx context.Context) error {
	// No server-side syncing needed here.
	return nil
}

func (r *Resolver) ResolveInteractive(ctx context.Context, reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	clientSecrets, err := r.lister.List(ctx)
	if err != nil {
		return err
	}

	for _, clientSecret := range clientSecrets {
		serverSecret, err := r.getter.Get(ctx, clientSecret.SecretName)
		if err != nil {
			return err
		}

		if serverSecret == nil || clientSecret.Timestamp >= serverSecret.Timestamp {
			fmt.Printf("Conflict for [%s]:\n", clientSecret.SecretName)
			if serverSecret == nil {
				fmt.Printf("Server version: <not found>\n")
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
				if err := r.saver.Save(ctx, clientSecret); err != nil {
					return err
				}
			case "2":
				// Keep server version, do nothing
			default:
				return errors.New("invalid version")
			}
		}
	}

	return nil
}
