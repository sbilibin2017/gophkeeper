package list

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/rsa"
	"github.com/sbilibin2017/gophkeeper/internal/transport/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/transport/http"
)

// RunHTTP lists and decrypts all secrets using HTTP facade.
func RunHTTP(ctx context.Context, token, serverURL string) (string, error) {
	client := http.New(serverURL)
	authFacade := facades.NewAuthHTTPFacade(client)
	secretReader := facades.NewSecretReadHTTPFacade(client)

	// Get username from token
	username, err := authFacade.GetUsername(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to get username: %w", err)
	}

	// Load RSA private key PEM and create decryptor
	_, privPEM, err := rsa.GetKeyPair(username)
	if err != nil {
		return "", fmt.Errorf("failed to load RSA key pair: %w", err)
	}
	decryptor, err := cryptor.New(cryptor.WithPrivateKeyPEM(privPEM))
	if err != nil {
		return "", fmt.Errorf("failed to create decryptor: %w", err)
	}

	// List all secrets for the user
	secrets, err := secretReader.List(ctx, username)
	if err != nil {
		return "", fmt.Errorf("failed to list secrets: %w", err)
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
			var bankcard models.SecretBankcard
			if err := json.Unmarshal(decrypted, &bankcard); err != nil {
				return "", fmt.Errorf("failed to unmarshal bankcard: %w", err)
			}
			out, _ := json.MarshalIndent(bankcard, "", "  ")
			builder.Write(out)

		case models.SecretTypeText:
			var text models.SecretText
			if err := json.Unmarshal(decrypted, &text); err != nil {
				return "", fmt.Errorf("failed to unmarshal text: %w", err)
			}
			out, _ := json.MarshalIndent(text, "", "  ")
			builder.Write(out)

		case models.SecretTypeBinary:
			var binary models.SecretBinary
			if err := json.Unmarshal(decrypted, &binary); err != nil {
				return "", fmt.Errorf("failed to unmarshal binary: %w", err)
			}
			out, _ := json.MarshalIndent(binary, "", "  ")
			builder.Write(out)

		case models.SecretTypeUser:
			var user models.SecretUser
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

// RunGRPC lists and decrypts all secrets using gRPC facade.
func RunGRPC(ctx context.Context, token, serverAddr string) (string, error) {
	client, err := grpc.New(serverAddr)
	if err != nil {
		return "", fmt.Errorf("failed to create gRPC client: %w", err)
	}
	defer client.Close()

	authFacade := facades.NewAuthGRPCFacade(client)
	secretReader := facades.NewSecretReadGRPCFacade(client)

	// Get username from token
	username, err := authFacade.GetUsername(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to get username: %w", err)
	}

	// Load RSA private key PEM and create decryptor
	_, privPEM, err := rsa.GetKeyPair(username)
	if err != nil {
		return "", fmt.Errorf("failed to load RSA key pair: %w", err)
	}
	decryptor, err := cryptor.New(cryptor.WithPrivateKeyPEM(privPEM))
	if err != nil {
		return "", fmt.Errorf("failed to create decryptor: %w", err)
	}

	// List all secrets for the user
	secrets, err := secretReader.List(ctx, username)
	if err != nil {
		return "", fmt.Errorf("failed to list secrets: %w", err)
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
			var bankcard models.SecretBankcard
			if err := json.Unmarshal(decrypted, &bankcard); err != nil {
				return "", fmt.Errorf("failed to unmarshal bankcard: %w", err)
			}
			out, _ := json.MarshalIndent(bankcard, "", "  ")
			builder.Write(out)

		case models.SecretTypeText:
			var text models.SecretText
			if err := json.Unmarshal(decrypted, &text); err != nil {
				return "", fmt.Errorf("failed to unmarshal text: %w", err)
			}
			out, _ := json.MarshalIndent(text, "", "  ")
			builder.Write(out)

		case models.SecretTypeBinary:
			var binary models.SecretBinary
			if err := json.Unmarshal(decrypted, &binary); err != nil {
				return "", fmt.Errorf("failed to unmarshal binary: %w", err)
			}
			out, _ := json.MarshalIndent(binary, "", "  ")
			builder.Write(out)

		case models.SecretTypeUser:
			var user models.SecretUser
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
