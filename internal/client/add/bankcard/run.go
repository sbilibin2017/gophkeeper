package bankcard

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

// RunHTTP encrypts bank card data and saves it via HTTP.
func RunHTTP(
	ctx context.Context,
	token string,
	serverURL string,
	secretName string,
	number string,
	owner string,
	exp string,
	cvv string,
	meta string,
) error {
	client := http.New(serverURL)
	authFacade := facades.NewAuthHTTPFacade(client)

	username, err := authFacade.GetUsername(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	userDB, err := db.New("sqlite", fmt.Sprintf("%s.db", username))
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer userDB.Close()

	pubPEM, _, err := rsa.GetKeyPair(username)
	if err != nil {
		return fmt.Errorf("failed to load RSA key pair: %w", err)
	}

	cr, err := cryptor.New(cryptor.WithPublicKeyPEM(pubPEM))
	if err != nil {
		return fmt.Errorf("failed to create cryptor: %w", err)
	}

	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}
	secret := models.SecretBankcard{
		Number: number,
		Owner:  owner,
		Exp:    exp,
		CVV:    cvv,
		Meta:   metaPtr,
	}

	plainBytes, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal bank card secret: %w", err)
	}

	encSecret, err := cr.Encrypt(plainBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt bank card: %w", err)
	}

	writeRepo := repositories.NewSecretWriteRepository(userDB)
	if err := writeRepo.Save(
		ctx,
		username,
		secretName,
		models.SecretTypeBankCard,
		encSecret.Ciphertext,
		encSecret.AESKeyEnc,
	); err != nil {
		return fmt.Errorf("failed to save encrypted bank card: %w", err)
	}

	return nil
}

// RunGRPC encrypts bank card data and saves it via gRPC.
func RunGRPC(
	ctx context.Context,
	token string,
	serverURL string,
	secretName string,
	number string,
	owner string,
	exp string,
	cvv string,
	meta string,
) error {
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

	authFacade := facades.NewAuthGRPCFacade(conn)

	username, err := authFacade.GetUsername(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get username via grpc facade: %w", err)
	}

	userDB, err := db.New("sqlite", fmt.Sprintf("%s.db", username))
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer userDB.Close()

	pubPEM, _, err := rsa.GetKeyPair(username)
	if err != nil {
		return fmt.Errorf("failed to load RSA key pair: %w", err)
	}

	cr, err := cryptor.New(cryptor.WithPublicKeyPEM(pubPEM))
	if err != nil {
		return fmt.Errorf("failed to create cryptor: %w", err)
	}

	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}
	secret := models.SecretBankcard{
		Number: number,
		Owner:  owner,
		Exp:    exp,
		CVV:    cvv,
		Meta:   metaPtr,
	}

	plainBytes, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal bank card secret: %w", err)
	}

	encSecret, err := cr.Encrypt(plainBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt bank card: %w", err)
	}

	writeRepo := repositories.NewSecretWriteRepository(userDB)
	if err := writeRepo.Save(
		ctx,
		username,
		secretName,
		models.SecretTypeBankCard,
		encSecret.Ciphertext,
		encSecret.AESKeyEnc,
	); err != nil {
		return fmt.Errorf("failed to save encrypted bank card: %w", err)
	}

	return nil
}
