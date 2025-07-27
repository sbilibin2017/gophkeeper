package apps

import (
	"fmt"
	"os"
	"time"

	"github.com/pressly/goose"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/db"
	"github.com/sbilibin2017/gophkeeper/inernal/cryptor"
	"github.com/sbilibin2017/gophkeeper/inernal/facades"
	"github.com/sbilibin2017/gophkeeper/inernal/repositories"
	"github.com/sbilibin2017/gophkeeper/inernal/usecases"
	"github.com/sbilibin2017/gophkeeper/inernal/validators"
)

// NewClientRegisterHTTPApp initializes and returns a ClientRegisterUsecase.
func NewClientRegisterHTTPApp(
	serverURL string,
	driverName string,
	databaseDSN string,
	pathToMigrationsDir string,
) (*usecases.ClientRegisterUsecase, error) {
	dbConn, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(1),
		db.WithConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}
	defer dbConn.Close()

	if err := goose.Up(dbConn.DB, pathToMigrationsDir); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	cli, err := http.New(serverURL, http.WithRetryPolicy(http.RetryPolicy{
		Count:   3,
		Wait:    time.Second,
		MaxWait: 5 * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	auth := facades.NewAuthHTTP(cli)

	usernameValidator := validators.NewUsernameValidator()
	passwordValidator := validators.NewPasswordValidator()

	uc := usecases.NewClientRegisterUsecase(usernameValidator, passwordValidator, auth)

	return uc, nil
}

// NewClientRegisterGRPCApp initializes and returns a ClientRegisterUsecase.
func NewClientRegisterGRPCApp(
	serverURL string,
	driverName string,
	databaseDSN string,
	pathToMigrationsDir string,
) (*usecases.ClientRegisterUsecase, error) {
	dbConn, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(1),
		db.WithConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}
	defer dbConn.Close()

	if err := goose.Up(dbConn.DB, pathToMigrationsDir); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	cli, err := grpc.New(serverURL, grpc.WithRetryPolicy(grpc.RetryPolicy{
		Count:   3,
		Wait:    time.Second,
		MaxWait: 5 * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	auth := facades.NewAuthGRPC(cli)

	usernameValidator := validators.NewUsernameValidator()
	passwordValidator := validators.NewPasswordValidator()

	uc := usecases.NewClientRegisterUsecase(usernameValidator, passwordValidator, auth)

	return uc, nil
}

// NewClientLoginApp initializes and returns a ClientLoginUsecase.
func NewClientLoginHTTPApp(
	serverURL string,
) (*usecases.ClientLoginUsecase, error) {
	cli, err := http.New(serverURL, http.WithRetryPolicy(http.RetryPolicy{
		Count:   3,
		Wait:    time.Second,
		MaxWait: 5 * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	auth := facades.NewAuthHTTP(cli)
	uc := usecases.NewClientLoginUsecase(auth)

	return uc, nil
}

// NewClientLoginGRPCApp initializes and returns a ClientLoginUsecase.
func NewClientLoginGRPCApp(
	serverURL string,
) (*usecases.ClientLoginUsecase, error) {
	cli, err := grpc.New(serverURL, grpc.WithRetryPolicy(grpc.RetryPolicy{
		Count:   3,
		Wait:    time.Second,
		MaxWait: 5 * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	auth := facades.NewAuthGRPC(cli)
	uc := usecases.NewClientLoginUsecase(auth)

	return uc, nil
}

// NewClientBankcardAddApp creates the usecase for adding encrypted bankcard secrets.
func NewClientBankcardAddApp(
	driverName string,
	databaseDSN string,
	pathToPublicKey string,
) (*usecases.ClientBankcardAddUsecase, error) {

	db, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(1),
		db.WithConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	publicKeyPEM, err := os.ReadFile(pathToPublicKey)
	if err != nil {
		return nil, err
	}

	cryptor, err := cryptor.New(
		cryptor.WithPublicKeyPEM(publicKeyPEM),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cryptor: %w", err)
	}

	saver := repositories.NewSecretWriteRepository(db)

	uc := usecases.NewClientBankcardAddUsecase(saver, cryptor)

	return uc, nil
}

// NewClientBinaryAddApp creates the usecase for adding encrypted binary secrets.
func NewClientBinaryAddApp(
	driverName string,
	databaseDSN string,
	pathToPublicKey string,
) (*usecases.ClientBinaryAddUsecase, error) {
	db, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(1),
		db.WithConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	publicKeyPEM, err := os.ReadFile(pathToPublicKey)
	if err != nil {
		return nil, err
	}

	cryptor, err := cryptor.New(
		cryptor.WithPublicKeyPEM(publicKeyPEM),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cryptor: %w", err)
	}

	saver := repositories.NewSecretWriteRepository(db)

	uc := usecases.NewClientBinaryAddUsecase(saver, cryptor)

	return uc, nil
}

// NewClientTextAddApp creates the usecase for adding encrypted text secrets.
func NewClientTextAddApp(
	driverName string,
	databaseDSN string,
	pathToPublicKey string,
) (*usecases.ClientTextAddUsecase, error) {
	db, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(1),
		db.WithConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	publicKeyPEM, err := os.ReadFile(pathToPublicKey)
	if err != nil {
		return nil, err
	}

	cryptor, err := cryptor.New(
		cryptor.WithPublicKeyPEM(publicKeyPEM),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cryptor: %w", err)
	}

	saver := repositories.NewSecretWriteRepository(db)

	uc := usecases.NewClientTextAddUsecase(saver, cryptor)

	return uc, nil
}

// NewClientUserAddApp creates the usecase for adding encrypted user/password secrets.
func NewClientUserAddApp(
	driverName string,
	databaseDSN string,
	pathToPublicKey string,
) (*usecases.ClientUserAddUsecase, error) {
	db, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(1),
		db.WithConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	publicKeyPEM, err := os.ReadFile(pathToPublicKey)
	if err != nil {
		return nil, err
	}

	cryptor, err := cryptor.New(
		cryptor.WithPublicKeyPEM(publicKeyPEM),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cryptor: %w", err)
	}

	saver := repositories.NewSecretWriteRepository(db)

	uc := usecases.NewClientUserAddUsecase(saver, cryptor)

	return uc, nil
}

// NewClientListHTTPApp creates the usecase for listing and decrypting secrets on client side.
func NewClientListHTTPApp(
	serverURL string,
	pathToPublicKey string,
) (*usecases.ClientListUsecase, error) {
	cli, err := http.New(serverURL, http.WithRetryPolicy(http.RetryPolicy{
		Count:   3,
		Wait:    time.Second,
		MaxWait: 5 * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	lister := facades.NewSecretReadHTTP(cli)

	publicKeyPEM, err := os.ReadFile(pathToPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	cryptor, err := cryptor.New(
		cryptor.WithPublicKeyPEM(publicKeyPEM),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cryptor: %w", err)
	}

	uc := usecases.NewClientListUsecase(lister, cryptor)

	return uc, nil
}

// NewClientListGRPCApp creates the usecase for listing and decrypting secrets on client side.
func NewClientListGRPCApp(
	serverURL string,
	pathToPublicKey string,
) (*usecases.ClientListUsecase, error) {
	cli, err := grpc.New(serverURL, grpc.WithRetryPolicy(grpc.RetryPolicy{
		Count:   3,
		Wait:    time.Second,
		MaxWait: 5 * time.Second,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	lister := facades.NewSecretReadGRPC(cli)

	publicKeyPEM, err := os.ReadFile(pathToPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	cryptor, err := cryptor.New(
		cryptor.WithPublicKeyPEM(publicKeyPEM),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cryptor: %w", err)
	}

	uc := usecases.NewClientListUsecase(lister, cryptor)

	return uc, nil
}

// NewClientSyncHTTPApp creates ClientSyncUsecase and builds HTTP client internally
func NewClientSyncHTTPApp(
	driverName string,
	databaseDSN string,
	serverURL string,
) (*usecases.ClientSyncUsecase, error) {

	dbConn, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(1),
		db.WithConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	httpClient, err := http.New(serverURL,
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    1 * time.Second,
			MaxWait: 5 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	clientLister := repositories.NewSecretReadRepository(dbConn)
	serverGetter := facades.NewSecretReadHTTP(httpClient)
	serverSaver := facades.NewSecretWriteHTTP(httpClient)

	uc := usecases.NewClientSyncUsecase(clientLister, serverGetter, serverSaver)

	return uc, nil
}

// NewClientSyncGRPCApp creates ClientSyncUsecase and builds gRPC client internally
func NewClientSyncGRPCApp(
	driverName string,
	databaseDSN string,
	serverURL string,
) (*usecases.ClientSyncUsecase, error) {

	dbConn, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(1),
		db.WithConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	conn, err := grpc.New(serverURL,
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    1 * time.Second,
			MaxWait: 5 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client connection: %w", err)
	}

	clientLister := repositories.NewSecretReadRepository(dbConn)
	serverGetter := facades.NewSecretReadGRPC(conn)
	serverSaver := facades.NewSecretWriteGRPC(conn)

	uc := usecases.NewClientSyncUsecase(clientLister, serverGetter, serverSaver)

	return uc, nil
}

// NewSyncInteractiveHTTPApp creates InteractiveSyncUsecase and builds HTTP client internally
func NewSyncInteractiveHTTPApp(
	driverName string,
	databaseDSN string,
	serverURL string,
	pathToPrivKey string,
) (*usecases.InteractiveSyncUsecase, error) {

	dbConn, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(1),
		db.WithConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	privKeyPEM, err := os.ReadFile(pathToPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	crypt, err := cryptor.New(cryptor.WithPrivateKeyPEM(privKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to create cryptor: %w", err)
	}

	httpClient, err := http.New(serverURL,
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    1 * time.Second,
			MaxWait: 5 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	clientLister := repositories.NewSecretReadRepository(dbConn)
	serverGetter := facades.NewSecretReadHTTP(httpClient)
	serverSaver := facades.NewSecretWriteHTTP(httpClient)

	uc := usecases.NewInteractiveSyncUsecase(clientLister, serverGetter, serverSaver, crypt)

	return uc, nil
}

// NewSyncInteractiveGRPCApp creates InteractiveSyncUsecase and builds gRPC client internally
func NewSyncInteractiveGRPCApp(
	driverName string,
	databaseDSN string,
	serverURL string,
	pathToPrivKeyFile string,
) (*usecases.InteractiveSyncUsecase, error) {

	dbConn, err := db.New(driverName, databaseDSN,
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(1),
		db.WithConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	privKeyPEM, err := os.ReadFile(pathToPrivKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	crypt, err := cryptor.New(cryptor.WithPublicKeyPEM(privKeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to create cryptor: %w", err)
	}

	conn, err := grpc.New(serverURL,
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    1 * time.Second,
			MaxWait: 5 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client connection: %w", err)
	}

	clientLister := repositories.NewSecretReadRepository(dbConn)
	serverGetter := facades.NewSecretReadGRPC(conn)
	serverSaver := facades.NewSecretWriteGRPC(conn)

	uc := usecases.NewInteractiveSyncUsecase(clientLister, serverGetter, serverSaver, crypt)

	return uc, nil
}

// NewServerSyncApp creates a new ServerSyncUsecase instance
func NewServerSyncApp() *usecases.ServerSyncUsecase {
	return usecases.NewServerSyncUsecase()
}
