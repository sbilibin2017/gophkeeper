package user

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
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// RunHTTP encrypts and saves a user secret using HTTP transport.
func RunHTTP(
	ctx context.Context,
	token string,
	serverURL string,
	secretName string,
	username string,
	password string,
	meta string,
) error {
	// 1. Get username from token
	client := http.New(serverURL)
	authFacade := facades.NewAuthHTTPFacade(client)

	usernameFromToken, err := authFacade.GetUsername(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	// 2. Open SQLite DB for this user
	userDB, err := db.New("sqlite", fmt.Sprintf("%s.db", usernameFromToken))
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer userDB.Close()

	// 3. Load RSA key pair
	pubPEM, _, err := rsa.GetKeyPair(usernameFromToken)
	if err != nil {
		return fmt.Errorf("failed to load RSA key pair: %w", err)
	}

	// 4. Create cryptor with public key
	cr, err := cryptor.New(cryptor.WithPublicKeyPEM(pubPEM))
	if err != nil {
		return fmt.Errorf("failed to create cryptor: %w", err)
	}

	// 5. Build SecretUser struct
	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}
	secret := models.SecretUser{
		Username: username,
		Password: password,
		Meta:     metaPtr,
	}

	// 6. Marshal to JSON
	plainBytes, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal user secret: %w", err)
	}

	// 7. Encrypt data
	encSecret, err := cr.Encrypt(plainBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt user secret: %w", err)
	}

	// 8. Save via repository (using new Save signature)
	writeRepo := repositories.NewSecretWriteRepository(userDB)
	if err := writeRepo.Save(ctx,
		usernameFromToken,
		secretName,
		models.SecretTypeUser,
		encSecret.Ciphertext,
		encSecret.AESKeyEnc,
	); err != nil {
		return fmt.Errorf("failed to save encrypted user secret: %w", err)
	}

	return nil
}

// RunGRPC encrypts and saves a user secret using gRPC transport.
func RunGRPC(
	ctx context.Context,
	token string,
	serverURL string,
	secretName string,
	username string,
	password string,
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

	// 2. Create gRPC auth client
	authClient := pb.NewAuthServiceClient(conn)

	// 3. Create context with token metadata
	md := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", token))
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	// 4. Call GetUsername
	resp, err := authClient.GetUsername(ctxWithToken, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("failed to get username via grpc: %w", err)
	}

	// 5. Open SQLite DB for user
	userDB, err := db.New("sqlite", fmt.Sprintf("%s.db", resp.Username))
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer userDB.Close()

	// 6. Load RSA key pair
	pubPEM, _, err := rsa.GetKeyPair(resp.Username)
	if err != nil {
		return fmt.Errorf("failed to load RSA key pair: %w", err)
	}

	// 7. Create cryptor with public key
	cr, err := cryptor.New(cryptor.WithPublicKeyPEM(pubPEM))
	if err != nil {
		return fmt.Errorf("failed to create cryptor: %w", err)
	}

	// 8. Build SecretUser struct
	var metaPtr *string
	if meta != "" {
		metaPtr = &meta
	}
	secret := models.SecretUser{
		Username: username,
		Password: password,
		Meta:     metaPtr,
	}

	// 9. Marshal to JSON
	plainBytes, err := json.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal user secret: %w", err)
	}

	// 10. Encrypt data
	encSecret, err := cr.Encrypt(plainBytes)
	if err != nil {
		return fmt.Errorf("failed to encrypt user secret: %w", err)
	}

	// 11. Save encrypted secret (using new Save signature)
	writeRepo := repositories.NewSecretWriteRepository(userDB)
	if err := writeRepo.Save(ctx,
		resp.Username,
		secretName,
		models.SecretTypeUser,
		encSecret.Ciphertext,
		encSecret.AESKeyEnc,
	); err != nil {
		return fmt.Errorf("failed to save encrypted user secret: %w", err)
	}

	return nil
}
