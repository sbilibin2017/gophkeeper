package resolver

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// Constants as you defined
const (
	ResolveStrategyServer      = "server"
	ResolveStrategyClient      = "client"
	ResolveStrategyInteractive = "interactive"
)

type ClientLister interface {
	List(ctx context.Context, secretOwner string) ([]*models.SecretDB, error)
}

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

type ServerGetter interface {
	Get(
		ctx context.Context,
		secretOwner string,
		secretName string,
		secretType string,
	) (*models.SecretDB, error)
}

type Decryptor interface {
	Decrypt(enc *models.SecretEncrypted) ([]byte, error)
}

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
		// FIXED ORDER: secretName, secretType
		serverSecret, err := sg.Get(ctx, secretOwner, clientSecret.SecretName, clientSecret.SecretType)
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
		// FIXED ORDER: secretName, secretType
		serverSecret, err := sg.Get(ctx, secretOwner, clientSecret.SecretName, clientSecret.SecretType)
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
				continue
			}

			serverPlain, err := d.Decrypt(&models.SecretEncrypted{
				Ciphertext: serverSecret.Ciphertext,
				AESKeyEnc:  serverSecret.AESKeyEnc,
			})
			if err != nil {
				continue
			}

			var clientPretty string
			var clientData any
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

			var serverPretty string
			var serverData any
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

			input := strings.TrimSpace(scanner.Text())

			if input != "1" && input != "2" {
				return errors.New("unsupported input")
			}

			if input == "1" {
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

			}
		}
	}

	return nil
}
