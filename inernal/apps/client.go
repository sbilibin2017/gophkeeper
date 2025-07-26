package apps

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"

	"github.com/sbilibin2017/gophkeeper/inernal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/cryptor"
	"github.com/sbilibin2017/gophkeeper/inernal/configs/db"
	"github.com/sbilibin2017/gophkeeper/inernal/repositories"

	"github.com/sbilibin2017/gophkeeper/inernal/facades"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	clientUsecases "github.com/sbilibin2017/gophkeeper/inernal/usecases/client"
	"github.com/sbilibin2017/gophkeeper/inernal/validators"
	grpcconn "google.golang.org/grpc"
)

// HTTP Register App
type ClientRegisterHTTPApp struct {
	registerUsecase *clientUsecases.RegisterUsecase
}

func NewClientRegisterHTTPApp(serverURL string) (*ClientRegisterHTTPApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	if err := goose.Up(dbConn.DB, "../../../migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	usernameValidator := validators.NewUsernameValidator()
	passwordValidator := validators.NewPasswordValidator()

	httpClient, err := http.New(serverURL)
	if err != nil {
		return nil, err
	}

	registerer, err := facades.NewAuthHTTPFacade(httpClient)
	if err != nil {
		return nil, err
	}

	registerUsecase, err := clientUsecases.NewRegisterUsecase(
		usernameValidator,
		passwordValidator,
		registerer,
	)
	if err != nil {
		return nil, err
	}

	return &ClientRegisterHTTPApp{registerUsecase: registerUsecase}, nil
}

func (a *ClientRegisterHTTPApp) Run(ctx context.Context, req *models.UserRegisterRequest) (*models.UserRegisterResponse, error) {
	return a.registerUsecase.Execute(ctx, req)
}

// GRPC Register App
type ClientRegisterGRPCApp struct {
	registerUsecase *clientUsecases.RegisterUsecase
}

func NewClientRegisterGRPCApp(serverURL string) (*ClientRegisterGRPCApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	if err := goose.Up(dbConn.DB, "../../../migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	usernameValidator := validators.NewUsernameValidator()
	passwordValidator := validators.NewPasswordValidator()

	grpcClient, err := grpc.New(serverURL)
	if err != nil {
		return nil, err
	}

	registerer, err := facades.NewAuthGRPCFacade(grpcClient)
	if err != nil {
		return nil, err
	}

	registerUsecase, err := clientUsecases.NewRegisterUsecase(
		usernameValidator,
		passwordValidator,
		registerer,
	)
	if err != nil {
		return nil, err
	}

	return &ClientRegisterGRPCApp{registerUsecase: registerUsecase}, nil
}

func (a *ClientRegisterGRPCApp) Run(ctx context.Context, req *models.UserRegisterRequest) (*models.UserRegisterResponse, error) {
	return a.registerUsecase.Execute(ctx, req)
}

// HTTP Login App
type ClientLoginHTTPApp struct {
	loginUsecase *clientUsecases.LoginUsecase
}

func NewClientLoginHTTPApp(serverURL string) (*ClientLoginHTTPApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	if err := goose.Up(dbConn.DB, "../../../migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	httpClient, err := http.New(serverURL)
	if err != nil {
		return nil, err
	}

	loginer, err := facades.NewAuthHTTPFacade(httpClient)
	if err != nil {
		return nil, err
	}

	loginUsecase, err := clientUsecases.NewLoginerUsecase(loginer)
	if err != nil {
		return nil, err
	}

	return &ClientLoginHTTPApp{loginUsecase: loginUsecase}, nil
}

func (a *ClientLoginHTTPApp) Run(ctx context.Context, req *models.UserLoginRequest) (*models.UserLoginResponse, error) {
	return a.loginUsecase.Execute(ctx, req)
}

// GRPC Login App
type ClientLoginGRPCApp struct {
	loginUsecase *clientUsecases.LoginUsecase
}

func NewClientLoginGRPCApp(serverURL string) (*ClientLoginGRPCApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	if err := goose.Up(dbConn.DB, "../../../migrations"); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	grpcClient, err := grpc.New(serverURL)
	if err != nil {
		return nil, err
	}

	loginer, err := facades.NewAuthGRPCFacade(grpcClient)
	if err != nil {
		return nil, err
	}

	loginUsecase, err := clientUsecases.NewLoginerUsecase(loginer)
	if err != nil {
		return nil, err
	}

	return &ClientLoginGRPCApp{loginUsecase: loginUsecase}, nil
}

func (a *ClientLoginGRPCApp) Run(ctx context.Context, req *models.UserLoginRequest) (*models.UserLoginResponse, error) {
	return a.loginUsecase.Execute(ctx, req)
}

type BankCardAddApp struct {
	usecase *clientUsecases.BankCardSecretAddUsecase
	db      *sqlx.DB
}

func NewBankCardAddApp(clientPubKeyFile string) (*BankCardAddApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	encryptor, err := cryptor.New(cryptor.WithPublicKeyFromCert(clientPubKeyFile))
	if err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	writer := repositories.NewSecretWriteRepository(dbConn)
	luhnValidator := validators.NewLuhnValidator()
	cvvValidator := validators.NewCVVValidator()

	usecase := clientUsecases.NewBankCardSecretAddUsecase(luhnValidator, cvvValidator, writer, encryptor)

	return &BankCardAddApp{usecase: usecase, db: dbConn}, nil
}

func (a *BankCardAddApp) Close() error {
	return a.db.Close()
}

func (a *BankCardAddApp) Run(ctx context.Context, secret *models.BankcardSecretAdd, token string) error {
	return a.usecase.Execute(ctx, secret, token)
}

type UserSecretAddApp struct {
	usecase *clientUsecases.UserSecretAddUsecase
	db      *sqlx.DB
}

func NewUserSecretAddApp(clientPubKeyFile string) (*UserSecretAddApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	encryptor, err := cryptor.New(cryptor.WithPublicKeyFromCert(clientPubKeyFile))
	if err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	writer := repositories.NewSecretWriteRepository(dbConn)
	usecase := clientUsecases.NewUserSecretAddUsecase(writer, encryptor)

	return &UserSecretAddApp{usecase: usecase, db: dbConn}, nil
}

func (a *UserSecretAddApp) Close() error {
	return a.db.Close()
}

func (a *UserSecretAddApp) Run(ctx context.Context, secret *models.UserSecretAdd, token string) error {
	return a.usecase.Execute(ctx, secret, token)
}

type BinarySecretAddApp struct {
	usecase *clientUsecases.BinarySecretAddUsecase
	db      *sqlx.DB
}

func NewBinarySecretAddApp(clientPubKeyFile string) (*BinarySecretAddApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	encryptor, err := cryptor.New(cryptor.WithPublicKeyFromCert(clientPubKeyFile))
	if err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	writer := repositories.NewSecretWriteRepository(dbConn)
	usecase := clientUsecases.NewBinarySecretAddUsecase(writer, encryptor)

	return &BinarySecretAddApp{usecase: usecase, db: dbConn}, nil
}

func (a *BinarySecretAddApp) Close() error {
	return a.db.Close()
}

func (a *BinarySecretAddApp) Run(ctx context.Context, secret *models.BinarySecretAdd, token string) error {
	return a.usecase.Execute(ctx, secret, token)
}

type TextSecretAddApp struct {
	usecase *clientUsecases.TextSecretAddUsecase
	db      *sqlx.DB
}

func NewTextSecretAddApp(clientPubKeyFile string) (*TextSecretAddApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	encryptor, err := cryptor.New(cryptor.WithPublicKeyFromCert(clientPubKeyFile))
	if err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	writer := repositories.NewSecretWriteRepository(dbConn)
	usecase := clientUsecases.NewTextSecretAddUsecase(writer, encryptor)

	return &TextSecretAddApp{usecase: usecase, db: dbConn}, nil
}

func (a *TextSecretAddApp) Close() error {
	return a.db.Close()
}

func (a *TextSecretAddApp) Run(ctx context.Context, secret *models.TextSecretAdd, token string) error {
	return a.usecase.Execute(ctx, secret, token)
}

// -------------------------------------------------------------

type ClientListHTTPApp struct {
	usecase    *clientUsecases.SecretClientListUsecase
	httpClient *resty.Client
}

func NewClientListHTTPApp(serverURL, privKeyPath string) (*ClientListHTTPApp, error) {
	encryptor, err := cryptor.New(
		cryptor.WithPrivateKeyFromFile(privKeyPath),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	httpClient, err := http.New(serverURL,
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    500 * time.Millisecond,
			MaxWait: 2 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	serverReader := facades.NewSecretHTTPReadFacade(httpClient)
	usecase := clientUsecases.NewSecretClientListUsecase(serverReader, encryptor)

	return &ClientListHTTPApp{
		usecase:    usecase,
		httpClient: httpClient,
	}, nil
}

func (c *ClientListHTTPApp) Run(ctx context.Context, req *models.SecretListRequest) (string, error) {
	return c.usecase.Execute(ctx, req)
}

type ClientListGRPCApp struct {
	usecase    *clientUsecases.SecretClientListUsecase
	grpcClient *grpcconn.ClientConn
}

func NewClientListGRPCApp(serverURL, privKeyPath string) (*ClientListGRPCApp, error) {
	encryptor, err := cryptor.New(
		cryptor.WithPrivateKeyFromFile(privKeyPath),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	grpcClient, err := grpc.New(serverURL,
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    500 * time.Millisecond,
			MaxWait: 2 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	serverReader := facades.NewSecretGRPCReadFacade(grpcClient)
	usecase := clientUsecases.NewSecretClientListUsecase(serverReader, encryptor)

	return &ClientListGRPCApp{
		usecase:    usecase,
		grpcClient: grpcClient,
	}, nil
}

func (c *ClientListGRPCApp) Run(ctx context.Context, req *models.SecretListRequest) (string, error) {
	return c.usecase.Execute(ctx, req)
}

func (c *ClientListGRPCApp) Close() error {
	if c.grpcClient != nil {
		return c.grpcClient.Close()
	}
	return nil
}

type ClientSyncHTTPApp struct {
	usecase *clientUsecases.ClientSyncUsecase
	db      *sqlx.DB
}

func NewClientSyncHTTPApp(serverURL string) (*ClientSyncHTTPApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	clientReader := repositories.NewSecretReadRepository(dbConn)

	httpClient, err := http.New(serverURL)
	if err != nil {
		dbConn.Close()
		return nil, err
	}

	serverReader := facades.NewSecretHTTPReadFacade(httpClient)
	serverWriter := facades.NewSecretHTTPWriteFacade(httpClient)

	usecase := clientUsecases.NewClientSyncUsecase(clientReader, serverReader, serverWriter)

	return &ClientSyncHTTPApp{
		usecase: usecase,
		db:      dbConn,
	}, nil
}

func (a *ClientSyncHTTPApp) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *ClientSyncHTTPApp) Run(ctx context.Context, token string) error {
	return a.usecase.Execute(ctx, token)
}

type ClientSyncGRPCApp struct {
	usecase    *clientUsecases.ClientSyncUsecase
	db         *sqlx.DB
	grpcClient *grpcconn.ClientConn
}

func NewClientSyncGRPCApp(serverURL string) (*ClientSyncGRPCApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	clientReader := repositories.NewSecretReadRepository(dbConn)

	grpcClient, err := grpc.New(serverURL)
	if err != nil {
		dbConn.Close()
		return nil, err
	}

	serverReader := facades.NewSecretGRPCReadFacade(grpcClient)
	serverWriter := facades.NewSecretGRPCWriteFacade(grpcClient)

	usecase := clientUsecases.NewClientSyncUsecase(clientReader, serverReader, serverWriter)

	return &ClientSyncGRPCApp{
		usecase:    usecase,
		db:         dbConn,
		grpcClient: grpcClient,
	}, nil
}

func (a *ClientSyncGRPCApp) Close() error {
	if a.grpcClient != nil {
		a.grpcClient.Close()
	}
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *ClientSyncGRPCApp) Run(ctx context.Context, token string) error {
	return a.usecase.Execute(ctx, token)
}

type ClientSyncServerApp struct {
	usecase *clientUsecases.ServerSyncUsecase
}

func NewClientSyncServerApp() *ClientSyncServerApp {
	return &ClientSyncServerApp{usecase: clientUsecases.NewServerSyncUsecase()}
}

func (a *ClientSyncServerApp) Execute(ctx context.Context, token string) error {
	return a.usecase.Execute(ctx, token)
}

type ClientSyncInteractiveHTTPApp struct {
	usecase *clientUsecases.InteractiveSyncUsecase
	db      *sqlx.DB
}

func NewClientSyncInteractiveHTTPApp(serverURL string) (*ClientSyncInteractiveHTTPApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	clientReader := repositories.NewSecretReadRepository(dbConn)

	httpClient, err := http.New(serverURL)
	if err != nil {
		dbConn.Close()
		return nil, err
	}

	serverReader := facades.NewSecretHTTPReadFacade(httpClient)
	serverWriter := facades.NewSecretHTTPWriteFacade(httpClient)

	decryptor, err := cryptor.New()
	if err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to create decryptor: %w", err)
	}

	usecase := clientUsecases.NewInteractiveSyncUsecase(clientReader, serverReader, serverWriter, decryptor)

	return &ClientSyncInteractiveHTTPApp{
		usecase: usecase,
		db:      dbConn,
	}, nil
}

func (a *ClientSyncInteractiveHTTPApp) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *ClientSyncInteractiveHTTPApp) Run(ctx context.Context, reader io.Reader, token string) error {
	return a.usecase.Execute(ctx, reader, token)
}

type ClientSyncInteractiveGRPCApp struct {
	usecase    *clientUsecases.InteractiveSyncUsecase
	db         *sqlx.DB
	grpcClient *grpcconn.ClientConn
}

func NewClientSyncInteractiveGRPCApp(serverURL string) (*ClientSyncInteractiveGRPCApp, error) {
	dbConn, err := db.New("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	clientReader := repositories.NewSecretReadRepository(dbConn)

	grpcClient, err := grpc.New(serverURL)
	if err != nil {
		dbConn.Close()
		return nil, err
	}

	serverReader := facades.NewSecretGRPCReadFacade(grpcClient)
	serverWriter := facades.NewSecretGRPCWriteFacade(grpcClient)

	decryptor, err := cryptor.New()
	if err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to create decryptor: %w", err)
	}

	usecase := clientUsecases.NewInteractiveSyncUsecase(clientReader, serverReader, serverWriter, decryptor)

	return &ClientSyncInteractiveGRPCApp{
		usecase:    usecase,
		db:         dbConn,
		grpcClient: grpcClient,
	}, nil
}

func (a *ClientSyncInteractiveGRPCApp) Close() error {
	if a.grpcClient != nil {
		a.grpcClient.Close()
	}
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func (a *ClientSyncInteractiveGRPCApp) Run(ctx context.Context, reader io.Reader, token string) error {
	return a.usecase.Execute(ctx, reader, token)
}
