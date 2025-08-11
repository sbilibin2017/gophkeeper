package text

import (
	"context"
	"encoding/json"
	"fmt"
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

// RunHTTP retrieves username from token, encrypts text secret data, and stores it in the DB.
func RunHTTP(
	ctx context.Context,
	token string,
	serverURL string,
	secretName string,
	data string,
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
	userDB, err := db.New(
		"sqlite",
		fmt.Sprintf("%s.db", username),
	)
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer userDB.Close()

	// 3. Load RSA key pair for this user
	pubPEM, _, err := rsa.GetKeyPair(username)
	if err != nil {
		return fmt.Errorf("failed to load RSA key pair: %w", err)
	}

	// 4. Create Cryptor with public key
	cr, err := cryptor.New(cryptor.WithPublicKeyPEM(pubPEM))
	if err != nil {
		return fmt.Errorf("failed to create cryptor: %w", err)
	}

	// 5. Build text secret struct
	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}
	secret := models.SecretText{
		Data: data,
		Meta: metaPtr,
	}

	// 6. Marshal text secret data to JSON
	plainBytes, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal text secret: %w", err)
	}

	// 7. Encrypt the text secret data
	encSecret, err := cr.Encrypt(plainBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt text secret: %w", err)
	}

	// 8. Save via repository using updated Save signature
	writeRepo := repositories.NewSecretWriteRepository(userDB)
	if err := writeRepo.Save(ctx,
		username,
		secretName,
		models.SecretTypeText,
		encSecret.Ciphertext,
		encSecret.AESKeyEnc,
	); err != nil {
		return fmt.Errorf("failed to save encrypted text secret: %w", err)
	}

	return nil
}

// RunGRPC retrieves username from token via gRPC, encrypts text secret data, and stores it in the DB.
func RunGRPC(
	ctx context.Context,
	token string,
	serverURL string,
	secretName string,
	data string,
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

	// 7. Prepare text secret struct
	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}
	secret := models.SecretText{
		Data: data,
		Meta: metaPtr,
	}

	// 8. Marshal text secret to JSON
	plainBytes, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal text secret: %w", err)
	}

	// 9. Encrypt text secret data
	encSecret, err := cr.Encrypt(plainBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt text secret: %w", err)
	}

	// 10. Save encrypted secret via repository using updated Save signature
	writeRepo := repositories.NewSecretWriteRepository(userDB)
	if err := writeRepo.Save(ctx,
		username,
		secretName,
		models.SecretTypeText,
		encSecret.Ciphertext,
		encSecret.AESKeyEnc,
	); err != nil {
		return fmt.Errorf("failed to save encrypted text secret: %w", err)
	}

	return nil
}
