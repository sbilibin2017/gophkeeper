package resolver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Secret is the constraint interface all secrets must implement
type Secret interface {
	GetSecretName() string
	GetUpdatedAt() time.Time
}

// Lister returns a list of client-side secrets
type Lister[T Secret] interface {
	List(ctx context.Context) ([]T, error)
}

// Getter retrieves a single server-side secret by name
type Getter[T Secret] interface {
	Get(ctx context.Context, secretName string) (T, error)
}

// Saver persists a secret (client overwrites server or vice versa)
type Saver[T Secret] interface {
	Save(ctx context.Context, secret T) error
}

// Resolver handles sync operations for secrets of type T.
type Resolver[T Secret] struct {
	lister Lister[T]
	getter Getter[T]
	saver  Saver[T]
	reader io.Reader
}

// NewResolver creates a new Resolver instance.
func NewResolver[T Secret](
	lister Lister[T],
	getter Getter[T],
	saver Saver[T],
	reader io.Reader,
) *Resolver[T] {
	return &Resolver[T]{
		lister: lister,
		getter: getter,
		saver:  saver,
		reader: reader,
	}
}

// Sync performs synchronization depending on the mode.
func (r *Resolver[T]) Resolve(ctx context.Context, mode string) error {
	switch mode {
	case models.SyncModeServer:
		return nil

	case models.SyncModeClient:
		clientSecrets, err := r.lister.List(ctx)
		if err != nil {
			return err
		}

		for _, clientSecret := range clientSecrets {
			serverSecret, err := r.getter.Get(ctx, clientSecret.GetSecretName())
			if err != nil {
				return err
			}

			if serverSecret.GetUpdatedAt().Before(clientSecret.GetUpdatedAt()) {
				if err := r.saver.Save(ctx, clientSecret); err != nil {
					return fmt.Errorf("failed to save client secret: %w", err)
				}
			}
		}
		return nil

	case models.SyncModeInteractive:
		scanner := bufio.NewScanner(r.reader)
		clientSecrets, err := r.lister.List(ctx)
		if err != nil {
			return err
		}

		for _, clientSecret := range clientSecrets {
			serverSecret, err := r.getter.Get(ctx, clientSecret.GetSecretName())
			if err != nil {
				return err
			}

			if !clientSecret.GetUpdatedAt().After(serverSecret.GetUpdatedAt()) {
				continue
			}

			fmt.Printf("Conflict for [%s]:\n", clientSecret.GetSecretName())
			fmt.Printf("1) Client version updated at %v\n", clientSecret.GetUpdatedAt())
			fmt.Printf("2) Server version updated at %v\n", serverSecret.GetUpdatedAt())
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
				continue
			default:
				return errors.New("invalid version")
			}
		}

		return nil

	default:
		return fmt.Errorf("unknown sync mode: %s", mode)
	}
}
