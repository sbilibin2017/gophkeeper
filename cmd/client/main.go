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

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/models/fields"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

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

	// convflictResolvingStrategy
	convflictResolvingStrategy string
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
	flag.StringVar(&secretType, "secret_type", "", "Type of the secret (e.g. bankcard, text, binary, user)")

	// Interactive conflict resolving
	flag.StringVar(&convflictResolvingStrategy, "interactive", "client", "Conflict resolving strategy(e.g. server/client/interactive)")
}

func main() {
	err := run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	flag.Parse()

	if len(os.Args) < 2 {
		return errors.New("no command specified")
	}

	protocol := scheme.GetSchemeFromURL(serverURL)
	if protocol == "" {
		return fmt.Errorf("unsupported or unknown protocol in server URL: %s", serverURL)
	}

	cmd := os.Args[1]

	switch cmd {
	case models.CommandRegister:
		switch protocol {
		case scheme.HTTP, scheme.HTTPS:
			token, err := registerHTTP(ctx, serverURL, certFile, certKey, username, password)
			if err != nil {
				return err
			}
			fmt.Println(token)
		case scheme.GRPC:
			token, err := registerGRPC(ctx, serverURL, certFile, certKey, username, password)
			if err != nil {
				return err
			}
			fmt.Println(token)
		default:
			return fmt.Errorf("unsupported protocol: %s", protocol)
		}
		return nil

	case models.CommandLogin:
		switch protocol {
		case scheme.HTTP, scheme.HTTPS:
			token, err := loginHTTP(ctx, serverURL, certFile, certKey, username, password)
			if err != nil {
				return err
			}
			fmt.Println(token)
		case scheme.GRPC:
			token, err := loginGRPC(ctx, serverURL, certFile, certKey, username, password)
			if err != nil {
				return err
			}
			fmt.Println(token)
		default:
			return fmt.Errorf("unsupported protocol: %s", protocol)
		}
		return nil

	case models.CommandLogout:
		switch protocol {
		case scheme.HTTP, scheme.HTTPS:
			if err := logoutHTTP(ctx, serverURL, certFile, certKey); err != nil {
				return err
			}
		case scheme.GRPC:
			if err := logoutGRPC(ctx, serverURL, certFile, certKey); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported protocol: %s", protocol)
		}
		return nil

	case models.CommandAddBankCard:
		err := addBankCardSecret(ctx, secretName, bankCardNumber, bankCardOwner, bankCardExp, bankCardCVV, meta)
		if err != nil {
			return err
		}
		return nil

	case models.CommandAddBinary:
		if err := addBinarySecret(ctx, secretName, binaryData, meta); err != nil {
			return err
		}
		return nil

	case models.CommandAddText:
		if err := addTextSecret(ctx, secretName, textContent, meta); err != nil {
			return err
		}
		return nil

	case models.CommandAddUser:
		if err := addUserSecret(ctx, secretName, username, password, meta); err != nil {
			return err
		}
		return nil

	case models.CommandGet:
		switch protocol {
		case scheme.HTTP, scheme.HTTPS:
			switch secretType {
			case models.SecretTypeBankCard:
				secret, err := getBankCardHTTP(ctx, serverURL, certFile, certKey, secretName)
				if err != nil {
					return err
				}
				data, err := json.MarshalIndent(secret, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal secret: %w", err)
				}
				fmt.Println(string(data))
			case models.SecretTypeUser:
				secret, err := getUserHTTP(ctx, serverURL, certFile, certKey, secretName)
				if err != nil {
					return err
				}
				data, err := json.MarshalIndent(secret, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal secret: %w", err)
				}
				fmt.Println(string(data))
			case models.SecretTypeText:
				secret, err := getTextHTTP(ctx, serverURL, certFile, certKey, secretName)
				if err != nil {
					return err
				}
				data, err := json.MarshalIndent(secret, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal secret: %w", err)
				}
				fmt.Println(string(data))
			case models.SecretTypeBinary:
				secret, err := getBinaryHTTP(ctx, serverURL, certFile, certKey, secretName)
				if err != nil {
					return err
				}
				data, err := json.MarshalIndent(secret, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal secret: %w", err)
				}
				fmt.Println(string(data))
			default:
				return fmt.Errorf("unsupported secret type for HTTP: %s", secretType)
			}

		case scheme.GRPC:
			switch secretType {
			case models.SecretTypeBankCard:
				secret, err := getBankCardGRPC(ctx, serverURL, certFile, certKey, secretName)
				if err != nil {
					return err
				}
				data, err := json.MarshalIndent(secret, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal secret: %w", err)
				}
				fmt.Println(string(data))
			case models.SecretTypeUser:
				secret, err := getUserGRPC(ctx, serverURL, certFile, certKey, secretName)
				if err != nil {
					return err
				}
				data, err := json.MarshalIndent(secret, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal secret: %w", err)
				}
				fmt.Println(string(data))
			case models.SecretTypeText:
				secret, err := getTextGRPC(ctx, serverURL, certFile, certKey, secretName)
				if err != nil {
					return err
				}
				data, err := json.MarshalIndent(secret, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal secret: %w", err)
				}
				fmt.Println(string(data))
			case models.SecretTypeBinary:
				secret, err := getBinaryGRPC(ctx, serverURL, certFile, certKey, secretName)
				if err != nil {
					return err
				}
				data, err := json.MarshalIndent(secret, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal secret: %w", err)
				}
				fmt.Println(string(data))
			default:
				return fmt.Errorf("unsupported secret type for gRPC: %s", secretType)
			}

		default:
			return fmt.Errorf("unsupported protocol: %s", protocol)
		}
		return nil

	case models.CommandSync:
		if convflictResolvingStrategy == "" {
			return fmt.Errorf("conflict resolving strategy is required")
		}

		reader := bufio.NewReader(os.Stdin)
		var err error

		switch protocol {
		case scheme.HTTP, scheme.HTTPS:
			switch secretType {
			case models.SecretTypeBankCard:
				err = syncBankCardHTTP(ctx, reader, convflictResolvingStrategy, serverURL, certFile, certKey, secretName)
			case models.SecretTypeUser:
				err = syncUserHTTP(ctx, reader, convflictResolvingStrategy, serverURL, certFile, certKey, secretName)
			case models.SecretTypeText:
				err = syncTextHTTP(ctx, reader, convflictResolvingStrategy, serverURL, certFile, certKey, secretName)
			case models.SecretTypeBinary:
				err = syncBinaryHTTP(ctx, reader, convflictResolvingStrategy, serverURL, certFile, certKey, secretName)
			default:
				return fmt.Errorf("unsupported secret type for sync over HTTP: %s", secretType)
			}
			if err != nil {
				return fmt.Errorf("%s HTTP sync failed: %w", secretType, err)
			}

		case scheme.GRPC:
			switch secretType {
			case models.SecretTypeBankCard:
				err = syncBankCardGRPC(ctx, reader, convflictResolvingStrategy, serverURL, certFile, certKey, secretName)
			case models.SecretTypeUser:
				err = syncUserGRPC(ctx, reader, convflictResolvingStrategy, serverURL, certFile, certKey, secretName)
			case models.SecretTypeText:
				err = syncTextGRPC(ctx, reader, convflictResolvingStrategy, serverURL, certFile, certKey, secretName)
			case models.SecretTypeBinary:
				err = syncBinaryGRPC(ctx, reader, convflictResolvingStrategy, serverURL, certFile, certKey, secretName)
			default:
				return fmt.Errorf("unsupported secret type for sync over gRPC: %s", secretType)
			}
			if err != nil {
				return fmt.Errorf("%s gRPC sync failed: %w", secretType, err)
			}

		default:
			return fmt.Errorf("unsupported protocol for sync: %s", protocol)
		}

		return nil

	case models.CommandVersion:
		fmt.Printf("Build version: %s\n", buildVersion)
		fmt.Printf("Build date: %s\n", buildDate)
		return nil

	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func registerHTTP(
	ctx context.Context,
	serverURL, certFile, certKey string,
	username, password string,
) (string, error) {
	req := &models.AuthRequest{
		Username: username,
		Password: password,
	}

	httpClient, err := http.New(serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP client: %w", err)
	}

	resp, err := client.RegisterHTTP(ctx, httpClient, req)
	if err != nil {
		return "", fmt.Errorf("register HTTP failed: %w", err)
	}

	return resp.Token, nil
}

func registerGRPC(
	ctx context.Context,
	serverURL, certFile, certKey string,
	username, password string,
) (string, error) {
	conn, err := grpc.New(serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewAuthServiceClient(conn)
	req := &models.AuthRequest{
		Username: username,
		Password: password,
	}

	resp, err := client.RegisterGRPC(ctx, grpcClient, req)
	if err != nil {
		return "", fmt.Errorf("register gRPC failed: %w", err)
	}

	return resp.Token, nil
}

func loginHTTP(
	ctx context.Context,
	serverURL, certFile, certKey string,
	username, password string,
) (string, error) {
	req := &models.AuthRequest{
		Username: username,
		Password: password,
	}

	httpClient, err := http.New(serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP client: %w", err)
	}

	resp, err := client.LoginHTTP(ctx, httpClient, req)
	if err != nil {
		return "", fmt.Errorf("login HTTP failed: %w", err)
	}

	return resp.Token, nil
}

func loginGRPC(
	ctx context.Context,
	serverURL, certFile, certKey string,
	username, password string,
) (string, error) {
	conn, err := grpc.New(serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewAuthServiceClient(conn)
	req := &models.AuthRequest{
		Username: username,
		Password: password,
	}

	resp, err := client.LoginGRPC(ctx, grpcClient, req)
	if err != nil {
		return "", fmt.Errorf("login gRPC failed: %w", err)
	}

	return resp.Token, nil
}

func logoutHTTP(
	ctx context.Context,
	serverURL, certFile, certKey string,
) error {
	httpClient, err := http.New(serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	if err := client.LogoutHTTP(ctx, httpClient); err != nil {
		return fmt.Errorf("logout HTTP failed: %w", err)
	}

	return nil
}
func logoutGRPC(
	ctx context.Context,
	serverURL, certFile, certKey string,
) error {
	conn, err := grpc.New(serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewAuthServiceClient(conn)

	if err := client.LogoutGRPC(ctx, grpcClient); err != nil {
		return fmt.Errorf("logout gRPC failed: %w", err)
	}

	return nil
}

func addBankCardSecret(
	ctx context.Context,
	secretName string,
	bankCardNumber string,
	bankCardOwner string,
	bankCardExp string,
	bankCardCVV string,
	meta string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	defer dbConn.Close()

	var metaFieldPtr *fields.StringMap
	if meta != "" {
		var metaMap map[string]string
		if err := json.Unmarshal([]byte(meta), &metaMap); err != nil {
			return fmt.Errorf("failed to parse meta JSON: %w", err)
		}
		sm := fields.StringMap{Map: metaMap}
		metaFieldPtr = &sm
	}

	req := &models.BankCardAddRequest{
		SecretName: secretName,
		Number:     bankCardNumber,
		Owner:      bankCardOwner,
		Exp:        bankCardExp,
		CVV:        bankCardCVV,
		Meta:       metaFieldPtr,
	}

	if err := client.BankCardAddClient(ctx, dbConn, req); err != nil {
		return fmt.Errorf("failed to add bank card: %w", err)
	}

	return nil
}

func addUserSecret(
	ctx context.Context,
	secretName, username, password, meta string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	defer dbConn.Close()

	var metaFieldPtr *fields.StringMap
	if meta != "" {
		var metaMap map[string]string
		if err := json.Unmarshal([]byte(meta), &metaMap); err != nil {
			return fmt.Errorf("failed to parse meta JSON: %w", err)
		}
		sm := fields.StringMap{Map: metaMap}
		metaFieldPtr = &sm
	}

	req := &models.UserAddRequest{
		SecretName: secretName,
		Username:   username,
		Password:   password,
		Meta:       metaFieldPtr,
	}

	if err := client.UserAddClient(ctx, dbConn, req); err != nil {
		return fmt.Errorf("failed to add user secret: %w", err)
	}

	return nil
}

func addTextSecret(
	ctx context.Context,
	secretName, content, meta string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	defer dbConn.Close()

	var metaFieldPtr *fields.StringMap
	if meta != "" {
		var metaMap map[string]string
		if err := json.Unmarshal([]byte(meta), &metaMap); err != nil {
			return fmt.Errorf("failed to parse meta JSON: %w", err)
		}
		sm := fields.StringMap{Map: metaMap}
		metaFieldPtr = &sm
	}

	req := &models.TextAddRequest{
		SecretName: secretName,
		Content:    content,
		Meta:       metaFieldPtr,
	}

	if err := client.TextAddClient(ctx, dbConn, req); err != nil {
		return fmt.Errorf("failed to add text content: %w", err)
	}

	return nil
}

func addBinarySecret(
	ctx context.Context,
	secretName string,
	binaryData string,
	meta string,
) error {
	decodedData, err := base64.StdEncoding.DecodeString(binaryData)
	if err != nil {
		return fmt.Errorf("failed to decode binary data (expected base64): %w", err)
	}

	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	defer dbConn.Close()

	var metaFieldPtr *fields.StringMap
	if meta != "" {
		var metaMap map[string]string
		if err := json.Unmarshal([]byte(meta), &metaMap); err != nil {
			return fmt.Errorf("failed to parse meta JSON into map[string]string: %w", err)
		}
		sm := fields.StringMap{Map: metaMap}
		metaFieldPtr = &sm
	}

	req := &models.BinaryAddRequest{
		SecretName: secretName,
		Data:       decodedData,
		Meta:       metaFieldPtr,
	}

	if err := client.BinaryAddClient(ctx, dbConn, req); err != nil {
		return fmt.Errorf("failed to add binary data: %w", err)
	}

	return nil
}

func getBankCardHTTP(
	ctx context.Context,
	serverURL, certFile, certKey string,
	secretName string,
) (*models.BankCardDB, error) {
	httpClient, err := http.New(serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	secret, err := client.BankCardGetHTTP(ctx, httpClient, secretName)
	if err != nil {
		return nil, fmt.Errorf("bank card GET HTTP failed: %w", err)
	}

	return secret, nil
}

func getBankCardGRPC(
	ctx context.Context,
	serverURL, certFile, certKey string,
	secretName string,
) (*models.BankCardDB, error) {
	conn, err := grpc.New(serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewBankCardServiceClient(conn)

	secret, err := client.BankCardGetGRPC(ctx, grpcClient, secretName)
	if err != nil {
		return nil, fmt.Errorf("bank card GET gRPC failed: %w", err)
	}

	return secret, nil
}

func getUserHTTP(
	ctx context.Context,
	serverURL, certFile, certKey string,
	secretName string,
) (*models.UserDB, error) {
	httpClient, err := http.New(serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	secret, err := client.UserGetHTTP(ctx, httpClient, secretName)
	if err != nil {
		return nil, fmt.Errorf("user GET HTTP failed: %w", err)
	}

	return secret, nil
}

func getUserGRPC(
	ctx context.Context,
	serverURL, certFile, certKey string,
	secretName string,
) (*models.UserDB, error) {
	if secretType == "" || secretName == "" {
		return nil, errors.New("usage: get -secret_type <type> -secret_name <name>")
	}

	conn, err := grpc.New(serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewUserServiceClient(conn)

	secret, err := client.UserGetGRPC(ctx, grpcClient, secretName)
	if err != nil {
		return nil, fmt.Errorf("user GET gRPC failed: %w", err)
	}

	return secret, nil
}

func getTextHTTP(
	ctx context.Context,
	serverURL, certFile, certKey string,
	secretName string,
) (*models.TextDB, error) {
	if secretType == "" || secretName == "" {
		return nil, errors.New("usage: get -secret_type <type> -secret_name <name>")
	}

	httpClient, err := http.New(serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	secret, err := client.TextGetHTTP(ctx, httpClient, secretName)
	if err != nil {
		return nil, fmt.Errorf("text GET HTTP failed: %w", err)
	}

	return secret, nil
}

func getTextGRPC(
	ctx context.Context,
	serverURL, certFile, certKey string,
	secretName string,
) (*models.TextDB, error) {
	if secretType == "" || secretName == "" {
		return nil, errors.New("usage: get -secret_type <type> -secret_name <name>")
	}

	conn, err := grpc.New(serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewTextServiceClient(conn)

	secret, err := client.TextGetGRPC(ctx, grpcClient, secretName)
	if err != nil {
		return nil, fmt.Errorf("text GET gRPC failed: %w", err)
	}

	return secret, nil
}

func getBinaryHTTP(
	ctx context.Context,
	serverURL, certFile, certKey string,
	secretName string,
) (*models.BinaryDB, error) {
	if secretType == "" || secretName == "" {
		return nil, errors.New("usage: get -secret_type <type> -secret_name <name>")
	}

	httpClient, err := http.New(serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	secret, err := client.BinaryGetHTTP(ctx, httpClient, secretName)
	if err != nil {
		return nil, fmt.Errorf("binary GET HTTP failed: %w", err)
	}

	return secret, nil
}

func getBinaryGRPC(
	ctx context.Context,
	serverURL, certFile, certKey string,
	secretName string,
) (*models.BinaryDB, error) {
	if secretType == "" || secretName == "" {
		return nil, errors.New("usage: get -secret_type <type> -secret_name <name>")
	}

	conn, err := grpc.New(serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewBinaryServiceClient(conn)

	secret, err := client.BinaryGetGRPC(ctx, grpcClient, secretName)
	if err != nil {
		return nil, fmt.Errorf("binary GET gRPC failed: %w", err)
	}

	return secret, nil
}

func syncBankCardHTTP(
	ctx context.Context,
	reader *bufio.Reader,
	strategy, serverURL, certFile, certKey, secretName string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	httpClient, err := http.New(
		serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	err = client.ResolveBankCardHTTP(
		ctx, reader, strategy,
		client.BankCardListClient,
		client.BankCardGetHTTP,
		client.BankCardAddHTTP,
		dbConn,
		httpClient,
		secretName,
	)
	if err != nil {
		return fmt.Errorf("bank card HTTP sync failed: %w", err)
	}
	return nil
}

func syncUserHTTP(
	ctx context.Context,
	reader *bufio.Reader,
	strategy, serverURL, certFile, certKey, secretName string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	httpClient, err := http.New(
		serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	err = client.ResolveUserHTTP(
		ctx, reader, strategy,
		client.UserListClient,
		client.UserGetHTTP,
		client.UserAddHTTP,
		dbConn,
		httpClient,
		secretName,
	)
	if err != nil {
		return fmt.Errorf("user HTTP sync failed: %w", err)
	}
	return nil
}

func syncTextHTTP(
	ctx context.Context,
	reader *bufio.Reader,
	strategy, serverURL, certFile, certKey, secretName string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	httpClient, err := http.New(
		serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	err = client.ResolveTextHTTP(
		ctx, reader, strategy,
		client.TextListClient,
		client.TextGetHTTP,
		client.TextAddHTTP,
		dbConn,
		httpClient,
		secretName,
	)
	if err != nil {
		return fmt.Errorf("text HTTP sync failed: %w", err)
	}
	return nil
}

func syncBinaryHTTP(
	ctx context.Context,
	reader *bufio.Reader,
	strategy, serverURL, certFile, certKey, secretName string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	httpClient, err := http.New(
		serverURL,
		http.WithTLSCert(http.TLSCert{CertFile: certFile, KeyFile: certKey}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	err = client.ResolveBinaryHTTP(
		ctx, reader, strategy,
		client.BinaryListClient,
		client.BinaryGetHTTP,
		client.BinaryAddHTTP,
		dbConn,
		httpClient,
		secretName,
	)
	if err != nil {
		return fmt.Errorf("binary HTTP sync failed: %w", err)
	}
	return nil
}

func syncBankCardGRPC(
	ctx context.Context,
	reader *bufio.Reader,
	strategy, serverURL, certFile, certKey, secretName string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	conn, err := grpc.New(
		serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewBankCardServiceClient(conn)
	err = client.ResolveBankCardGRPC(
		ctx, reader, strategy,
		client.BankCardListClient,
		client.BankCardGetGRPC,
		client.BankCardAddGRPC,
		dbConn,
		grpcClient,
		secretName,
	)
	if err != nil {
		return fmt.Errorf("bank card gRPC sync failed: %w", err)
	}
	return nil
}

func syncUserGRPC(
	ctx context.Context,
	reader *bufio.Reader,
	strategy, serverURL, certFile, certKey, secretName string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	conn, err := grpc.New(
		serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewUserServiceClient(conn)
	err = client.ResolveUserGRPC(
		ctx, reader, strategy,
		client.UserListClient,
		client.UserGetGRPC,
		client.UserAddGRPC,
		dbConn,
		grpcClient,
		secretName,
	)
	if err != nil {
		return fmt.Errorf("user gRPC sync failed: %w", err)
	}
	return nil
}

func syncTextGRPC(
	ctx context.Context,
	reader *bufio.Reader,
	strategy, serverURL, certFile, certKey, secretName string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	conn, err := grpc.New(
		serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewTextServiceClient(conn)
	err = client.ResolveTextGRPC(
		ctx, reader, strategy,
		client.TextListClient,
		client.TextGetGRPC,
		client.TextAddGRPC,
		dbConn,
		grpcClient,
		secretName,
	)
	if err != nil {
		return fmt.Errorf("text gRPC sync failed: %w", err)
	}
	return nil
}

func syncBinaryGRPC(
	ctx context.Context,
	reader *bufio.Reader,
	strategy, serverURL, certFile, certKey, secretName string,
) error {
	dbConn, err := db.NewDB("sqlite", "gophkeeper.db",
		db.WithMaxOpenConns(1),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	conn, err := grpc.New(
		serverURL,
		grpc.WithTLSCert(grpc.TLSCert{CertFile: certFile, KeyFile: certKey}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    100 * time.Millisecond,
			MaxWait: 1 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	grpcClient := pb.NewBinaryServiceClient(conn)
	err = client.ResolveBinaryGRPC(
		ctx, reader, strategy,
		client.BinaryListClient,
		client.BinaryGetGRPC,
		client.BinaryAddGRPC,
		dbConn,
		grpcClient,
		secretName,
	)
	if err != nil {
		return fmt.Errorf("binary gRPC sync failed: %w", err)
	}
	return nil
}
