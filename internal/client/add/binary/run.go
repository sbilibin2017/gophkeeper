package binary

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/db"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/rsa"
	"github.com/sbilibin2017/gophkeeper/internal/transport/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/transport/http"
)

// RunHTTP reads binary data from file, encrypts, and saves it via HTTP.
func RunHTTP(
	ctx context.Context,
	token string,
	serverURL string,
	secretName string,
	dataPath string,
	meta string,
) error {
	// 1. Get username from token
	client := http.New(serverURL)
	authFacade := facades.NewAuthHTTPFacade(client)

	username, err := authFacade.GetUsername(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	// 2. Open SQLite DB for this user
	userDB, err := db.New("sqlite", fmt.Sprintf("%s.db", username))
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer userDB.Close()

	// 3. Load RSA key pair for user
	pubPEM, _, err := rsa.GetKeyPair(username)
	if err != nil {
		return fmt.Errorf("failed to load RSA key pair: %w", err)
	}

	// 4. Create Cryptor with public key
	cr, err := cryptor.New(cryptor.WithPublicKeyPEM(pubPEM))
	if err != nil {
		return fmt.Errorf("failed to create cryptor: %w", err)
	}

	// 5. Read binary data from file
	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		return fmt.Errorf("failed to read binary file: %w", err)
	}

	// 6. Build secret struct
	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}
	secret := models.SecretBinary{
		Data: dataBytes,
		Meta: metaPtr,
	}

	// 7. Marshal secret to JSON
	plainBytes, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal binary secret: %w", err)
	}

	// 8. Encrypt data
	encSecret, err := cr.Encrypt(plainBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt binary secret: %w", err)
	}

	// 9. Save encrypted secret via repository using updated Save signature
	writeRepo := repositories.NewSecretWriteRepository(userDB)
	if err := writeRepo.Save(ctx,
		username,
		secretName,
		models.SecretTypeBinary,
		encSecret.Ciphertext,
		encSecret.AESKeyEnc,
	); err != nil {
		return fmt.Errorf("failed to save encrypted binary secret: %w", err)
	}

	return nil
}

// RunGRPC reads binary data from file, encrypts, and saves it via gRPC.
func RunGRPC(
	ctx context.Context,
	token string,
	serverURL string,
	secretName string,
	dataPath string,
	meta string,
) error {
	// 1. Connect to gRPC server with retry policy
	conn, err := grpc.New(
		serverURL,
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    500 * time.Millisecond,
			MaxWait: 2 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create grpc client connection: %w", err)
	}
	defer conn.Close()

	// 2. Create AuthGRPCFacade instance
	authFacade := facades.NewAuthGRPCFacade(conn)

	// 3. Use facade to get username from token
	username, err := authFacade.GetUsername(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get username via grpc facade: %w", err)
	}

	// 4. Open SQLite DB for this user
	userDB, err := db.New("sqlite", fmt.Sprintf("%s.db", username))
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer userDB.Close()

	// 5. Load RSA key pair for user
	pubPEM, _, err := rsa.GetKeyPair(username)
	if err != nil {
		return fmt.Errorf("failed to load RSA key pair: %w", err)
	}

	// 6. Create Cryptor with public key
	cr, err := cryptor.New(cryptor.WithPublicKeyPEM(pubPEM))
	if err != nil {
		return fmt.Errorf("failed to create cryptor: %w", err)
	}

	// 7. Read binary data from file
	dataBytes, err := os.ReadFile(dataPath)
	if err != nil {
		return fmt.Errorf("failed to read binary file: %w", err)
	}

	// 8. Build secret struct
	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}
	secret := models.SecretBinary{
		Data: dataBytes,
		Meta: metaPtr,
	}

	// 9. Marshal secret to JSON
	plainBytes, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal binary secret: %w", err)
	}

	// 10. Encrypt data
	encSecret, err := cr.Encrypt(plainBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt binary secret: %w", err)
	}

	// 11. Save encrypted secret via repository using updated Save signature
	writeRepo := repositories.NewSecretWriteRepository(userDB)
	if err := writeRepo.Save(ctx,
		username,
		secretName,
		models.SecretTypeBinary,
		encSecret.Ciphertext,
		encSecret.AESKeyEnc,
	); err != nil {
		return fmt.Errorf("failed to save encrypted binary secret: %w", err)
	}

	return nil
}
