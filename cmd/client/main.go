package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/models/fields"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func main() {
	flag.Parse()
	err := run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

var (
	// Build info
	buildVersion = "N/A"
	buildDate    = "N/A"

	// Config
	serverURL string
	certFile  string
	certKey   string

	// Shared
	secretName string
	meta       string

	// AuthRequest / UserAddRequest
	username string
	password string

	// BankCardRequest
	bankCardNumber string
	bankCardOwner  string
	bankCardExp    string
	bankCardCVV    string

	// TextAddRequest
	textContent string

	// BinaryAddRequest
	binaryData string

	// Secret type for filtering
	secretType string

	// Conflict resolving strategy
	conflictResolvingStrategy string
)

func init() {
	// Shared
	flag.StringVar(&secretName, "secret_name", "", "Unique name of the secret (used for add/filter requests)")
	flag.StringVar(&meta, "meta", "", "Additional metadata or notes")

	// AuthRequest / UserAddRequest
	flag.StringVar(&username, "username", "", "Username (auth or user secret)")
	flag.StringVar(&password, "password", "", "Password (auth or user secret)")

	// BankCardRequest
	flag.StringVar(&bankCardNumber, "number", "", "Card number (Luhn validated)")
	flag.StringVar(&bankCardOwner, "owner", "", "Card owner name")
	flag.StringVar(&bankCardExp, "exp", "", "Card expiration date (MM/YY)")
	flag.StringVar(&bankCardCVV, "cvv", "", "Card CVV code (3 digits)")

	// BinaryAddRequest
	flag.StringVar(&binaryData, "data", "", "Base64-encoded binary data")

	// TextAddRequest
	flag.StringVar(&textContent, "content", "", "Text content to store")

	// Config
	flag.StringVar(&serverURL, "server_url", "http://localhost:8080", "Server base URL")
	flag.StringVar(&certFile, "cert_file", "", "TLS certificate file path")
	flag.StringVar(&certKey, "cert_key", "", "TLS certificate key file path")

	// Secret type
	flag.StringVar(&secretType, "secret_type", "", fmt.Sprintf("Type of the secret (e.g. %s, %s, %s, %s)",
		models.SecretTypeBankCard, models.SecretTypeText, models.SecretTypeBinary, models.SecretTypeUser))

	// Conflict resolving
	flag.StringVar(&conflictResolvingStrategy, "conflict_resolving_strategy", "client", "Conflict resolving strategy (e.g. server/client/interactive)")
}

func run(ctx context.Context) error {
	// Require a command line argument
	if len(os.Args) < 2 {
		return errors.New("no command specified")
	}

	// Parse protocol scheme from serverURL
	protocol := scheme.GetSchemeFromURL(serverURL)
	cmd := os.Args[1]
	reader := bufio.NewReader(os.Stdin)

	// Open database connection with some settings
	dbConn, err := db.NewDB("sqlite", "client.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	switch cmd {
	case models.CommandRegister:
		switch protocol {
		case scheme.HTTP, scheme.HTTPS:
			httpClient, err := http.New(serverURL,
				http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}
			token, err := registerHTTP(ctx, httpClient, username, password)
			if err != nil {
				return err
			}
			fmt.Println(token)
			return nil

		case scheme.GRPC:
			grpcConn, err := grpc.New(serverURL,
				grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}
			defer grpcConn.Close()

			grpcClient := pb.NewAuthServiceClient(grpcConn)
			token, err := registerGRPC(ctx, grpcClient, username, password)
			if err != nil {
				return err
			}
			fmt.Println(token)
			return nil

		default:
			return fmt.Errorf("unsupported protocol for register command: %s", protocol)
		}

	case models.CommandLogin:
		switch protocol {
		case scheme.HTTP, scheme.HTTPS:
			httpClient, err := http.New(serverURL,
				http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}
			token, err := loginHTTP(ctx, httpClient, username, password)
			if err != nil {
				return err
			}
			fmt.Println(token)
			return nil

		case scheme.GRPC:
			grpcConn, err := grpc.New(serverURL,
				grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}
			defer grpcConn.Close()

			grpcClient := pb.NewAuthServiceClient(grpcConn)
			token, err := loginGRPC(ctx, grpcClient, username, password)
			if err != nil {
				return err
			}
			fmt.Println(token)
			return nil

		default:
			return fmt.Errorf("unsupported protocol for login command: %s", protocol)
		}

	case models.CommandLogout:
		switch protocol {
		case scheme.HTTP, scheme.HTTPS:
			httpClient, err := http.New(serverURL,
				http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}
			return logoutHTTP(ctx, httpClient)

		case scheme.GRPC:
			grpcConn, err := grpc.New(serverURL,
				grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}
			defer grpcConn.Close()

			grpcClient := pb.NewAuthServiceClient(grpcConn)
			return logoutGRPC(ctx, grpcClient)

		default:
			return fmt.Errorf("unsupported protocol for logout command: %s", protocol)
		}

	case models.CommandAddBankCard:
		err := addBankCardClient(ctx, dbConn,
			secretName,
			bankCardNumber,
			bankCardOwner,
			bankCardExp,
			bankCardCVV,
			meta,
		)
		if err != nil {
			return err
		}
		fmt.Println("Bank card added successfully")
		return nil

	case models.CommandAddBinary:
		err := addBinaryClient(ctx, dbConn,
			secretName,
			binaryData,
			meta,
		)
		if err != nil {
			return err
		}
		fmt.Println("Binary data added successfully")
		return nil

	case models.CommandAddText:
		err := addTextClient(ctx, dbConn,
			secretName,
			textContent,
			meta,
		)
		if err != nil {
			return err
		}
		fmt.Println("Text content added successfully")
		return nil

	case models.CommandAddUser:
		err := addUserClient(ctx, dbConn,
			secretName,
			username,
			password,
			meta,
		)
		if err != nil {
			return err
		}
		fmt.Println("User secret added successfully")
		return nil

	case models.CommandGet:
		switch protocol {
		case scheme.HTTP, scheme.HTTPS:
			httpClient, err := http.New(serverURL,
				http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}

			var result string
			switch secretType {
			case models.SecretTypeBankCard:
				result, err = getBankCardHTTP(ctx, httpClient, secretName)
			case models.SecretTypeUser:
				result, err = getUserHTTP(ctx, httpClient, secretName)
			case models.SecretTypeText:
				result, err = getTextHTTP(ctx, httpClient, secretName)
			case models.SecretTypeBinary:
				result, err = getBinaryHTTP(ctx, httpClient, secretName)
			default:
				return fmt.Errorf("unknown secret_type: %s", secretType)
			}
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil

		case scheme.GRPC:
			grpcConn, err := grpc.New(serverURL,
				grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}
			defer grpcConn.Close()

			var result string
			switch secretType {
			case models.SecretTypeBankCard:
				grpcClient := pb.NewBankCardServiceClient(grpcConn)
				result, err = getBankCardGRPC(ctx, grpcClient, secretName)
			case models.SecretTypeUser:
				grpcClient := pb.NewUserServiceClient(grpcConn)
				result, err = getUserGRPC(ctx, grpcClient, secretName)
			case models.SecretTypeText:
				grpcClient := pb.NewTextServiceClient(grpcConn)
				result, err = getTextGRPC(ctx, grpcClient, secretName)
			case models.SecretTypeBinary:
				grpcClient := pb.NewBinaryServiceClient(grpcConn)
				result, err = getBinaryGRPC(ctx, grpcClient, secretName)
			default:
				return fmt.Errorf("unknown secret_type: %s", secretType)
			}
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil

		default:
			var result string
			switch secretType {
			case models.SecretTypeBankCard:
				result, err = getBankCardClient(ctx, dbConn, secretName)
			case models.SecretTypeUser:
				result, err = getUserClient(ctx, dbConn, secretName)
			case models.SecretTypeText:
				result, err = getTextClient(ctx, dbConn, secretName)
			case models.SecretTypeBinary:
				result, err = getBinaryClient(ctx, dbConn, secretName)
			default:
				return fmt.Errorf("unknown secret_type: %s", secretType)
			}
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil
		}

	case models.CommandList:
		switch protocol {
		case scheme.HTTP, scheme.HTTPS:
			httpClient, err := http.New(serverURL,
				http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}

			var result string
			switch secretType {
			case models.SecretTypeBankCard:
				result, err = listBankCardsHTTP(ctx, httpClient)
			case models.SecretTypeUser:
				result, err = listUsersHTTP(ctx, httpClient)
			case models.SecretTypeText:
				result, err = listTextsHTTP(ctx, httpClient)
			case models.SecretTypeBinary:
				result, err = listBinariesHTTP(ctx, httpClient)
			default:
				result, err = listAllHTTP(ctx, httpClient)
			}
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil

		case scheme.GRPC:
			grpcConn, err := grpc.New(serverURL,
				grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}
			defer grpcConn.Close()

			var result string
			switch secretType {
			case models.SecretTypeBankCard:
				result, err = listBankCardsGRPC(ctx, pb.NewBankCardServiceClient(grpcConn))
			case models.SecretTypeUser:
				result, err = listUsersGRPC(ctx, pb.NewUserServiceClient(grpcConn))
			case models.SecretTypeText:
				result, err = listTextsGRPC(ctx, pb.NewTextServiceClient(grpcConn))
			case models.SecretTypeBinary:
				result, err = listBinariesGRPC(ctx, pb.NewBinaryServiceClient(grpcConn))
			default:
				result, err = listAllGRPC(ctx,
					pb.NewBankCardServiceClient(grpcConn),
					pb.NewUserServiceClient(grpcConn),
					pb.NewTextServiceClient(grpcConn),
					pb.NewBinaryServiceClient(grpcConn),
				)
			}
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil

		default:
			var result string
			switch secretType {
			case models.SecretTypeBankCard:
				result, err = listBankCardsClient(ctx, dbConn)
			case models.SecretTypeUser:
				result, err = listUsersClient(ctx, dbConn)
			case models.SecretTypeText:
				result, err = listTextsClient(ctx, dbConn)
			case models.SecretTypeBinary:
				result, err = listBinariesClient(ctx, dbConn)
			default:
				result, err = listAllClient(ctx, dbConn)
			}
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil
		}

	case models.CommandSync:
		switch protocol {
		case scheme.GRPC:
			grpcConn, err := grpc.New(serverURL,
				grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}
			defer grpcConn.Close()

			return syncGRPC(
				ctx,
				dbConn,
				pb.NewBinaryServiceClient(grpcConn),
				pb.NewBankCardServiceClient(grpcConn),
				pb.NewTextServiceClient(grpcConn),
				pb.NewUserServiceClient(grpcConn),
				reader,
				conflictResolvingStrategy,
			)

		case scheme.HTTP, scheme.HTTPS:
			httpClient, err := http.New(serverURL,
				http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
			)
			if err != nil {
				return err
			}

			return syncHTTP(
				ctx,
				dbConn,
				httpClient,
				httpClient,
				httpClient,
				httpClient,
				reader,
				conflictResolvingStrategy,
			)

		default:
			return fmt.Errorf("unsupported protocol for sync: %s", protocol)
		}

	case models.CommandVersion:
		fmt.Printf("Build version: %s\n", buildVersion)
		fmt.Printf("Build date: %s\n", buildDate)
		return nil

	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

// syncGRPC synchronizes all secret types via gRPC using the provided clients.
func syncGRPC(
	ctx context.Context,
	db *sqlx.DB,
	binaryClient pb.BinaryServiceClient,
	bankCardClient pb.BankCardServiceClient,
	textClient pb.TextServiceClient,
	userClient pb.UserServiceClient,
	reader *bufio.Reader,
	strategy string,
) error {
	// Sync Binary secrets
	if err := client.ResolveBinaryGRPC(
		ctx,
		reader,
		strategy,
		client.BinaryListClient,
		client.BinaryGetGRPC,
		client.BinaryAddGRPC,
		db,
		binaryClient,
	); err != nil {
		return err
	}

	// Sync BankCard secrets
	if err := client.ResolveBankCardGRPC(
		ctx,
		reader,
		strategy,
		client.BankCardListClient,
		client.BankCardGetGRPC,
		client.BankCardAddGRPC,
		db,
		bankCardClient,
	); err != nil {
		return err
	}

	// Sync Text secrets
	if err := client.ResolveTextGRPC(
		ctx,
		reader,
		strategy,
		client.TextListClient,
		client.TextGetGRPC,
		client.TextAddGRPC,
		db,
		textClient,
	); err != nil {
		return err
	}

	// Sync User secrets
	if err := client.ResolveUserGRPC(
		ctx,
		reader,
		strategy,
		client.UserListClient,
		client.UserGetGRPC,
		client.UserAddGRPC,
		db,
		userClient,
	); err != nil {
		return err
	}

	return nil
}

// syncHTTP synchronizes all secret types via HTTP using the provided REST clients.
func syncHTTP(
	ctx context.Context,
	db *sqlx.DB,
	binaryClient *resty.Client,
	bankCardClient *resty.Client,
	textClient *resty.Client,
	userClient *resty.Client,
	reader *bufio.Reader,
	strategy string,
) error {
	// Sync Binary secrets
	if err := client.ResolveBinaryHTTP(
		ctx,
		reader,
		strategy,
		client.BinaryListClient,
		client.BinaryGetHTTP,
		client.BinaryAddHTTP,
		db,
		binaryClient,
	); err != nil {
		return err
	}

	// Sync BankCard secrets
	if err := client.ResolveBankCardHTTP(
		ctx,
		reader,
		strategy,
		client.BankCardListClient,
		client.BankCardGetHTTP,
		client.BankCardAddHTTP,
		db,
		bankCardClient,
	); err != nil {
		return err
	}

	// Sync Text secrets
	if err := client.ResolveTextHTTP(
		ctx,
		reader,
		strategy,
		client.TextListClient,
		client.TextGetHTTP,
		client.TextAddHTTP,
		db,
		textClient,
	); err != nil {
		return err
	}

	// Sync User secrets
	if err := client.ResolveUserHTTP(
		ctx,
		reader,
		strategy,
		client.UserListClient,
		client.UserGetHTTP,
		client.UserAddHTTP,
		db,
		userClient,
	); err != nil {
		return err
	}

	return nil
}

func listBankCardsHTTP(ctx context.Context, httpClient *resty.Client) (string, error) {
	secrets, err := client.BankCardListHTTP(ctx, httpClient)
	if err != nil {
		return "", fmt.Errorf("bank card LIST HTTP failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listUsersHTTP(ctx context.Context, httpClient *resty.Client) (string, error) {
	secrets, err := client.UserListHTTP(ctx, httpClient)
	if err != nil {
		return "", fmt.Errorf("user LIST HTTP failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listTextsHTTP(ctx context.Context, httpClient *resty.Client) (string, error) {
	secrets, err := client.TextListHTTP(ctx, httpClient)
	if err != nil {
		return "", fmt.Errorf("text LIST HTTP failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listBinariesHTTP(ctx context.Context, httpClient *resty.Client) (string, error) {
	secrets, err := client.BinaryListHTTP(ctx, httpClient)
	if err != nil {
		return "", fmt.Errorf("binary LIST HTTP failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// gRPC List functions returning JSON string (private)

func listBankCardsGRPC(ctx context.Context, grpcClient pb.BankCardServiceClient) (string, error) {
	secrets, err := client.BankCardListGRPC(ctx, grpcClient)
	if err != nil {
		return "", fmt.Errorf("bank card LIST gRPC failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listUsersGRPC(ctx context.Context, grpcClient pb.UserServiceClient) (string, error) {
	secrets, err := client.UserListGRPC(ctx, grpcClient)
	if err != nil {
		return "", fmt.Errorf("user LIST gRPC failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listTextsGRPC(ctx context.Context, grpcClient pb.TextServiceClient) (string, error) {
	secrets, err := client.TextListGRPC(ctx, grpcClient)
	if err != nil {
		return "", fmt.Errorf("text LIST gRPC failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listBinariesGRPC(ctx context.Context, grpcClient pb.BinaryServiceClient) (string, error) {
	secrets, err := client.BinaryListGRPC(ctx, grpcClient)
	if err != nil {
		return "", fmt.Errorf("binary LIST gRPC failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Client (DB) List functions returning JSON string (private)

func listBankCardsClient(ctx context.Context, db *sqlx.DB) (string, error) {
	secrets, err := client.BankCardListClient(ctx, db)
	if err != nil {
		return "", fmt.Errorf("bank card LIST client failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listUsersClient(ctx context.Context, db *sqlx.DB) (string, error) {
	secrets, err := client.UserListClient(ctx, db)
	if err != nil {
		return "", fmt.Errorf("user LIST client failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listTextsClient(ctx context.Context, db *sqlx.DB) (string, error) {
	secrets, err := client.TextListClient(ctx, db)
	if err != nil {
		return "", fmt.Errorf("text LIST client failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listBinariesClient(ctx context.Context, db *sqlx.DB) (string, error) {
	secrets, err := client.BinaryListClient(ctx, db)
	if err != nil {
		return "", fmt.Errorf("binary LIST client failed: %w", err)
	}
	b, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ListAll functions returning combined JSON string (private)

func listAllGRPC(
	ctx context.Context,
	bankCardClient pb.BankCardServiceClient,
	userClient pb.UserServiceClient,
	textClient pb.TextServiceClient,
	binaryClient pb.BinaryServiceClient,
) (string, error) {
	var allSecrets []any

	bankCards, err := client.BankCardListGRPC(ctx, bankCardClient)
	if err != nil {
		return "", fmt.Errorf("bank card LIST gRPC failed: %w", err)
	}
	for _, c := range bankCards {
		allSecrets = append(allSecrets, c)
	}

	users, err := client.UserListGRPC(ctx, userClient)
	if err != nil {
		return "", fmt.Errorf("user LIST gRPC failed: %w", err)
	}
	for _, u := range users {
		allSecrets = append(allSecrets, u)
	}

	texts, err := client.TextListGRPC(ctx, textClient)
	if err != nil {
		return "", fmt.Errorf("text LIST gRPC failed: %w", err)
	}
	for _, t := range texts {
		allSecrets = append(allSecrets, t)
	}

	binaries, err := client.BinaryListGRPC(ctx, binaryClient)
	if err != nil {
		return "", fmt.Errorf("binary LIST gRPC failed: %w", err)
	}
	for _, b := range binaries {
		allSecrets = append(allSecrets, b)
	}

	b, err := json.MarshalIndent(allSecrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listAllHTTP(
	ctx context.Context,
	httpClient *resty.Client,
) (string, error) {
	var allSecrets []any

	bankCards, err := client.BankCardListHTTP(ctx, httpClient)
	if err != nil {
		return "", fmt.Errorf("bank card LIST HTTP failed: %w", err)
	}
	for _, c := range bankCards {
		allSecrets = append(allSecrets, c)
	}

	users, err := client.UserListHTTP(ctx, httpClient)
	if err != nil {
		return "", fmt.Errorf("user LIST HTTP failed: %w", err)
	}
	for _, u := range users {
		allSecrets = append(allSecrets, u)
	}

	texts, err := client.TextListHTTP(ctx, httpClient)
	if err != nil {
		return "", fmt.Errorf("text LIST HTTP failed: %w", err)
	}
	for _, t := range texts {
		allSecrets = append(allSecrets, t)
	}

	binaries, err := client.BinaryListHTTP(ctx, httpClient)
	if err != nil {
		return "", fmt.Errorf("binary LIST HTTP failed: %w", err)
	}
	for _, b := range binaries {
		allSecrets = append(allSecrets, b)
	}

	b, err := json.MarshalIndent(allSecrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listAllClient(
	ctx context.Context,
	db *sqlx.DB,
) (string, error) {
	var allSecrets []any

	bankCards, err := client.BankCardListClient(ctx, db)
	if err != nil {
		return "", fmt.Errorf("bank card LIST client failed: %w", err)
	}
	for _, c := range bankCards {
		allSecrets = append(allSecrets, c)
	}

	users, err := client.UserListClient(ctx, db)
	if err != nil {
		return "", fmt.Errorf("user LIST client failed: %w", err)
	}
	for _, u := range users {
		allSecrets = append(allSecrets, u)
	}

	texts, err := client.TextListClient(ctx, db)
	if err != nil {
		return "", fmt.Errorf("text LIST client failed: %w", err)
	}
	for _, t := range texts {
		allSecrets = append(allSecrets, t)
	}

	binaries, err := client.BinaryListClient(ctx, db)
	if err != nil {
		return "", fmt.Errorf("binary LIST client failed: %w", err)
	}
	for _, b := range binaries {
		allSecrets = append(allSecrets, b)
	}

	b, err := json.MarshalIndent(allSecrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// HTTP Getters returning JSON string (private)

func getBankCardHTTP(ctx context.Context, httpClient *resty.Client, secretName string) (string, error) {
	secret, err := client.BankCardGetHTTP(ctx, httpClient, secretName)
	if err != nil {
		return "", fmt.Errorf("bank card GET HTTP failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getUserHTTP(ctx context.Context, httpClient *resty.Client, secretName string) (string, error) {
	secret, err := client.UserGetHTTP(ctx, httpClient, secretName)
	if err != nil {
		return "", fmt.Errorf("user GET HTTP failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getTextHTTP(ctx context.Context, httpClient *resty.Client, secretName string) (string, error) {
	secret, err := client.TextGetHTTP(ctx, httpClient, secretName)
	if err != nil {
		return "", fmt.Errorf("text GET HTTP failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getBinaryHTTP(ctx context.Context, httpClient *resty.Client, secretName string) (string, error) {
	secret, err := client.BinaryGetHTTP(ctx, httpClient, secretName)
	if err != nil {
		return "", fmt.Errorf("binary GET HTTP failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// gRPC Getters returning JSON string (private)

func getBankCardGRPC(ctx context.Context, grpcClient pb.BankCardServiceClient, secretName string) (string, error) {
	secret, err := client.BankCardGetGRPC(ctx, grpcClient, secretName)
	if err != nil {
		return "", fmt.Errorf("bank card GET gRPC failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getUserGRPC(ctx context.Context, grpcClient pb.UserServiceClient, secretName string) (string, error) {
	secret, err := client.UserGetGRPC(ctx, grpcClient, secretName)
	if err != nil {
		return "", fmt.Errorf("user GET gRPC failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getTextGRPC(ctx context.Context, grpcClient pb.TextServiceClient, secretName string) (string, error) {
	secret, err := client.TextGetGRPC(ctx, grpcClient, secretName)
	if err != nil {
		return "", fmt.Errorf("text GET gRPC failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getBinaryGRPC(ctx context.Context, grpcClient pb.BinaryServiceClient, secretName string) (string, error) {
	secret, err := client.BinaryGetGRPC(ctx, grpcClient, secretName)
	if err != nil {
		return "", fmt.Errorf("binary GET gRPC failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Client (DB) Getters returning JSON string (private)

func getBankCardClient(ctx context.Context, db *sqlx.DB, secretName string) (string, error) {
	secret, err := client.BankCardGetClient(ctx, db, secretName)
	if err != nil {
		return "", fmt.Errorf("bank card GET client failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getUserClient(ctx context.Context, db *sqlx.DB, secretName string) (string, error) {
	secret, err := client.UserGetClient(ctx, db, secretName)
	if err != nil {
		return "", fmt.Errorf("user GET client failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getTextClient(ctx context.Context, db *sqlx.DB, secretName string) (string, error) {
	secret, err := client.TextGetClient(ctx, db, secretName)
	if err != nil {
		return "", fmt.Errorf("text GET client failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getBinaryClient(ctx context.Context, db *sqlx.DB, secretName string) (string, error) {
	secret, err := client.BinaryGetClient(ctx, db, secretName)
	if err != nil {
		return "", fmt.Errorf("binary GET client failed: %w", err)
	}
	b, err := json.MarshalIndent(secret, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// RegisterHTTP registers a user via HTTP and returns the token string.
func registerHTTP(
	ctx context.Context,
	httpClient *resty.Client,
	username, password string,
) (string, error) {
	req := &models.AuthRequest{
		Username: username,
		Password: password,
	}
	resp, err := client.RegisterHTTP(ctx, httpClient, req)
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}

// RegisterGRPC registers a user via gRPC and returns the token string.
func registerGRPC(
	ctx context.Context,
	grpcClient pb.AuthServiceClient,
	username, password string,
) (string, error) {
	req := &models.AuthRequest{
		Username: username,
		Password: password,
	}
	resp, err := client.RegisterGRPC(ctx, grpcClient, req)
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}

// LoginHTTP logs in via HTTP and returns the token string.
func loginHTTP(
	ctx context.Context,
	httpClient *resty.Client,
	username, password string,
) (string, error) {
	req := &models.AuthRequest{
		Username: username,
		Password: password,
	}
	resp, err := client.LoginHTTP(ctx, httpClient, req)
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}

// LoginGRPC logs in via gRPC and returns the token string.
func loginGRPC(
	ctx context.Context,
	grpcClient pb.AuthServiceClient,
	username, password string,
) (string, error) {
	req := &models.AuthRequest{
		Username: username,
		Password: password,
	}
	resp, err := client.LoginGRPC(ctx, grpcClient, req)
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}

func logoutHTTP(
	ctx context.Context,
	httpClient *resty.Client,
) error {
	return client.LogoutHTTP(ctx, httpClient)
}

func logoutGRPC(
	ctx context.Context,
	grpcClient pb.AuthServiceClient,
) error {
	return client.LogoutGRPC(ctx, grpcClient)
}

func addBinaryClient(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
	binaryData string,
	meta string,
) error {
	decodedData, err := base64.StdEncoding.DecodeString(binaryData)
	if err != nil {
		return fmt.Errorf("failed to decode binary data (expected base64): %w", err)
	}

	metaFieldPtr, err := fields.ParseMeta(meta)
	if err != nil {
		return err
	}

	req := &models.BinaryAddRequest{
		SecretName: secretName,
		Data:       decodedData,
		Meta:       metaFieldPtr,
	}

	if err := client.BinaryAddClient(ctx, db, req); err != nil {
		return fmt.Errorf("failed to add binary data: %w", err)
	}

	return nil
}

func addBankCardClient(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
	number string,
	owner string,
	exp string,
	cvv string,
	meta string,
) error {
	metaFieldPtr, err := fields.ParseMeta(meta)
	if err != nil {
		return err
	}

	req := &models.BankCardAddRequest{
		SecretName: secretName,
		Number:     number,
		Owner:      owner,
		Exp:        exp,
		CVV:        cvv,
		Meta:       metaFieldPtr,
	}

	if err := client.BankCardAddClient(ctx, db, req); err != nil {
		return fmt.Errorf("failed to add bank card: %w", err)
	}

	return nil
}

func addUserClient(
	ctx context.Context,
	db *sqlx.DB,
	secretName, username, password, meta string,
) error {
	metaFieldPtr, err := fields.ParseMeta(meta)
	if err != nil {
		return err
	}

	req := &models.UserAddRequest{
		SecretName: secretName,
		Username:   username,
		Password:   password,
		Meta:       metaFieldPtr,
	}

	if err := client.UserAddClient(ctx, db, req); err != nil {
		return fmt.Errorf("failed to add user secret: %w", err)
	}

	return nil
}

func addTextClient(
	ctx context.Context,
	db *sqlx.DB,
	secretName, content, meta string,
) error {
	metaFieldPtr, err := fields.ParseMeta(meta)
	if err != nil {
		return err
	}

	req := &models.TextAddRequest{
		SecretName: secretName,
		Content:    content,
		Meta:       metaFieldPtr,
	}

	if err := client.TextAddClient(ctx, db, req); err != nil {
		return fmt.Errorf("failed to add text content: %w", err)
	}

	return nil
}
