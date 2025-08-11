package sync

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/db"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/resolver"
	"github.com/sbilibin2017/gophkeeper/internal/rsa"
	"github.com/sbilibin2017/gophkeeper/internal/transport/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/transport/http"
)

// RunHTTP synchronizes bank card secrets using the specified conflict resolution mode.
func RunHTTP(
	ctx context.Context,
	token string,
	serverURL string,
	mode string,
) error {
	// Create resty client with retry policy
	restyClient := http.New(serverURL,
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    500 * time.Millisecond,
			MaxWait: 2 * time.Second,
		}),
	)

	authFacade := facades.NewAuthHTTPFacade(restyClient)

	username, err := authFacade.GetUsername(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	userDB, err := db.New("sqlite", fmt.Sprintf("%s.db", username))
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer userDB.Close()

	_, privPEM, err := rsa.GetKeyPair(username)
	if err != nil {
		return fmt.Errorf("failed to load RSA key pair: %w", err)
	}

	decryptor, err := cryptor.New(cryptor.WithPrivateKeyPEM(privPEM))
	if err != nil {
		return fmt.Errorf("failed to create decryptor: %w", err)
	}

	clientLister := facades.NewSecretReadHTTPFacade(restyClient)
	serverGetter := repositories.NewSecretReadRepository(userDB)
	serverSaverLocal := repositories.NewSecretWriteRepository(userDB)

	switch mode {
	case models.SyncModeClient:
		err = resolver.ClientSyncClient(ctx, clientLister, serverGetter, serverSaverLocal, username)
		if err != nil {
			return fmt.Errorf("failed to synchronize secrets with client resolution: %w", err)
		}

	case models.SyncModeInteractive:
		err = resolver.ClientSyncInteractive(ctx, clientLister, serverGetter, serverSaverLocal, decryptor, username, os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to synchronize secrets interactively: %w", err)
		}

	default:
		return fmt.Errorf("unsupported sync mode: %s", mode)
	}

	return nil
}

// RunGRPC synchronizes secrets using gRPC transport with the specified conflict resolution mode.
func RunGRPC(
	ctx context.Context,
	token string,
	serverAddr string,
	mode string,
) error {
	// Create gRPC connection with retry policy
	conn, err := grpc.New(serverAddr,
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    500 * time.Millisecond,
			MaxWait: 2 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create grpc connection: %w", err)
	}
	defer conn.Close()

	authFacade := facades.NewAuthGRPCFacade(conn)

	username, err := authFacade.GetUsername(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	userDB, err := db.New("sqlite", fmt.Sprintf("%s.db", username))
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	defer userDB.Close()

	_, privPEM, err := rsa.GetKeyPair(username)
	if err != nil {
		return fmt.Errorf("failed to load RSA key pair: %w", err)
	}

	decryptor, err := cryptor.New(cryptor.WithPrivateKeyPEM(privPEM))
	if err != nil {
		return fmt.Errorf("failed to create decryptor: %w", err)
	}

	clientLister := facades.NewSecretReadGRPCFacade(conn)
	serverGetter := repositories.NewSecretReadRepository(userDB)
	serverSaverLocal := repositories.NewSecretWriteRepository(userDB)

	switch mode {
	case models.SyncModeClient:
		err = resolver.ClientSyncClient(ctx, clientLister, serverGetter, serverSaverLocal, username)
		if err != nil {
			return fmt.Errorf("failed to synchronize secrets with client resolution: %w", err)
		}

	case models.SyncModeInteractive:
		err = resolver.ClientSyncInteractive(ctx, clientLister, serverGetter, serverSaverLocal, decryptor, username, os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to synchronize secrets interactively: %w", err)
		}

	default:
		return fmt.Errorf("unsupported sync mode: %s", mode)
	}

	return nil
}
