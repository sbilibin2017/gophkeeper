package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pressly/goose"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/db"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/transport/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/transport/http"
	"github.com/sbilibin2017/gophkeeper/internal/validators"
)

func main() {
	flag.Parse()
	err := run(context.Background(), os.Args)
	if err != nil {
		help := client.GetHelp()
		log.Print(help)
		log.Fatal(err)
	}
}

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

var (
	serverURL string
	pubKey    string
	privKey   string
	token     string

	secretType string
	secretName string

	number string
	owner  string
	exp    string
	cvv    string

	data string

	username string
	password string

	meta string

	syncMode string
)

func init() {
	flag.StringVar(&serverURL, "server-url", "", "Server URL")
	flag.StringVar(&pubKey, "pubkey", "", "Public key")
	flag.StringVar(&privKey, "privkey", "", "Private key")
	flag.StringVar(&token, "token", "", "Authentication token")

	flag.StringVar(&secretType, "secret-type", "", "Type of secret: bankcard, text, binary, user")
	flag.StringVar(&secretName, "secret-name", "", "Secret name")

	flag.StringVar(&number, "number", "", "Bankcard number")
	flag.StringVar(&owner, "owner", "", "Bankcard owner")
	flag.StringVar(&exp, "exp", "", "Bankcard expiry date")
	flag.StringVar(&cvv, "cvv", "", "Bankcard CVV")

	flag.StringVar(&data, "data", "", "Text data")

	flag.StringVar(&username, "username", "", "Username")
	flag.StringVar(&password, "password", "", "Password")

	flag.StringVar(&meta, "meta", "", "Optional meta")

	flag.StringVar(&syncMode, "sync-mode", "", "Sync mode")
}

// run executes the client command specified in args.
// It supports commands: register, login, add secrets (bankcard, text, binary, user),
// synchronize secrets with the server, show version info, and help.
// Depending on the command and server URL scheme (HTTP(S)/gRPC), it creates
// appropriate connections and clients, handling encryption and retries.
func run(ctx context.Context, args []string) error {
	command := client.GetCommand(args)

	schm := scheme.GetSchemeFromURL(serverURL)

	switch command {
	case client.CommandRegister:
		if err := validators.ValidateUsername(username); err != nil {
			return fmt.Errorf("invalid username: %w", err)
		}
		if err := validators.ValidatePassword(password); err != nil {
			return fmt.Errorf("invalid password: %w", err)
		}

		switch schm {
		case scheme.HTTP, scheme.HTTPS:
			dbConn, err := db.New(
				"sqlite",
				"client.db",
				db.WithMaxOpenConns(1),
				db.WithMaxIdleConns(1),
				db.WithConnMaxLifetime(30*time.Minute),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to DB: %w", err)
			}
			defer dbConn.Close()

			if err := goose.SetDialect("sqlite"); err != nil {
				return fmt.Errorf("failed to set goose dialect: %w", err)
			}

			if err := goose.Up(dbConn.DB, "../../../migrations"); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			httpClient, err := http.New(serverURL, http.WithRetryPolicy(http.RetryPolicy{
				Count:   3,
				Wait:    1 * time.Second,
				MaxWait: 5 * time.Second,
			}))
			if err != nil {
				return err
			}
			authFacade := facades.NewAuthHTTPFacade(httpClient)

			tk, err := client.ClientRegister(ctx, authFacade, username, password)
			if err != nil {
				return err
			}
			fmt.Println(tk)
			return nil

		case scheme.GRPC:
			grpcConn, err := grpc.New(serverURL, grpc.WithRetryPolicy(grpc.RetryPolicy{
				Count:   3,
				Wait:    1 * time.Second,
				MaxWait: 5 * time.Second,
			}))
			if err != nil {
				return err
			}
			defer grpcConn.Close()

			authFacade := facades.NewAuthGRPCFacade(grpcConn)

			tk, err := client.ClientRegister(ctx, authFacade, username, password)
			if err != nil {
				return err
			}
			fmt.Println(tk)
			return nil

		default:
			return errors.New("unsupported scheme")
		}

	case client.CommandLogin:
		switch schm {
		case scheme.HTTP, scheme.HTTPS:
			httpClient, err := http.New(serverURL, http.WithRetryPolicy(http.RetryPolicy{
				Count:   3,
				Wait:    1 * time.Second,
				MaxWait: 5 * time.Second,
			}))
			if err != nil {
				return err
			}
			authFacade := facades.NewAuthHTTPFacade(httpClient)

			tk, err := client.ClientLogin(ctx, authFacade, username, password)
			if err != nil {
				return err
			}
			fmt.Println(tk)
			return nil

		case scheme.GRPC:
			grpcConn, err := grpc.New(serverURL, grpc.WithRetryPolicy(grpc.RetryPolicy{
				Count:   3,
				Wait:    1 * time.Second,
				MaxWait: 5 * time.Second,
			}))
			if err != nil {
				return err
			}
			defer grpcConn.Close()

			authFacade := facades.NewAuthGRPCFacade(grpcConn)

			tk, err := client.ClientLogin(ctx, authFacade, username, password)
			if err != nil {
				return err
			}
			fmt.Println(tk)
			return nil

		default:
			return errors.New("unsupported scheme")
		}

	case client.CommandAddBankcard:
		if err := validators.ValidateLuhn(number); err != nil {
			return fmt.Errorf("invalid card number: %w", err)
		}
		if err := validators.ValidateCVV(cvv); err != nil {
			return fmt.Errorf("invalid CVV: %w", err)
		}

		dbConn, err := db.New("sqlite", "client.db",
			db.WithMaxOpenConns(1),
			db.WithMaxIdleConns(1),
			db.WithConnMaxLifetime(30*time.Minute),
		)
		if err != nil {
			return fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer dbConn.Close()

		clientWriter := repositories.NewSecretWriteRepository(dbConn)

		cryptorInst, err := cryptor.New(
			cryptor.WithPublicKeyPEM([]byte(pubKey)),
			cryptor.WithPrivateKeyPEM([]byte(privKey)),
		)
		if err != nil {
			return fmt.Errorf("cryptor setup failed: %w", err)
		}

		return client.ClientAddBankcard(ctx, clientWriter, cryptorInst, token, secretName, number, owner, exp, cvv, meta)

	case client.CommandAddText:
		dbConn, err := db.New("sqlite", "client.db",
			db.WithMaxOpenConns(1),
			db.WithMaxIdleConns(1),
			db.WithConnMaxLifetime(30*time.Minute),
		)
		if err != nil {
			return fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer dbConn.Close()

		clientWriter := repositories.NewSecretWriteRepository(dbConn)

		cryptorInst, err := cryptor.New(
			cryptor.WithPublicKeyPEM([]byte(pubKey)),
			cryptor.WithPrivateKeyPEM([]byte(privKey)),
		)
		if err != nil {
			return fmt.Errorf("cryptor setup failed: %w", err)
		}

		return client.ClientAddText(ctx, clientWriter, cryptorInst, token, secretName, data, meta)

	case client.CommandAddBinary:
		dbConn, err := db.New("sqlite", "client.db",
			db.WithMaxOpenConns(1),
			db.WithMaxIdleConns(1),
			db.WithConnMaxLifetime(30*time.Minute),
		)
		if err != nil {
			return fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer dbConn.Close()

		clientWriter := repositories.NewSecretWriteRepository(dbConn)

		cryptorInst, err := cryptor.New(
			cryptor.WithPublicKeyPEM([]byte(pubKey)),
			cryptor.WithPrivateKeyPEM([]byte(privKey)),
		)
		if err != nil {
			return fmt.Errorf("cryptor setup failed: %w", err)
		}

		return client.ClientAddBinary(ctx, clientWriter, cryptorInst, token, secretName, data, meta)

	case client.CommandAddUser:
		dbConn, err := db.New("sqlite", "client.db",
			db.WithMaxOpenConns(1),
			db.WithMaxIdleConns(1),
			db.WithConnMaxLifetime(30*time.Minute),
		)
		if err != nil {
			return fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer dbConn.Close()

		clientWriter := repositories.NewSecretWriteRepository(dbConn)

		cryptorInst, err := cryptor.New(
			cryptor.WithPublicKeyPEM([]byte(pubKey)),
			cryptor.WithPrivateKeyPEM([]byte(privKey)),
		)
		if err != nil {
			return fmt.Errorf("cryptor setup failed: %w", err)
		}

		return client.ClientAddUser(ctx, clientWriter, cryptorInst, token, secretName, username, password, meta)

	case client.CommandSync:
		dbConn, err := db.New("sqlite", "client.db",
			db.WithMaxOpenConns(1),
			db.WithMaxIdleConns(1),
			db.WithConnMaxLifetime(30*time.Minute),
		)
		if err != nil {
			return fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer dbConn.Close()

		clientReader := repositories.NewSecretReadRepository(dbConn)

		cryptorInst, err := cryptor.New(
			cryptor.WithPublicKeyPEM([]byte(pubKey)),
			cryptor.WithPrivateKeyPEM([]byte(privKey)),
		)
		if err != nil {
			return fmt.Errorf("cryptor setup failed: %w", err)
		}

		switch schm {
		case scheme.HTTP, scheme.HTTPS:
			httpClient, err := http.New(serverURL, http.WithRetryPolicy(http.RetryPolicy{
				Count:   3,
				Wait:    1 * time.Second,
				MaxWait: 5 * time.Second,
			}))
			if err != nil {
				return err
			}
			serverGetter := facades.NewSecretReaderHTTP(httpClient)
			serverSaver := facades.NewSecretWriterHTTP(httpClient)

			switch syncMode {
			case client.ResolveStrategyServer:
				return nil

			case client.ResolveStrategyClient:
				if err := client.ClientSyncClient(ctx, clientReader, serverGetter, serverSaver, token); err != nil {
					return fmt.Errorf("client sync failed: %w", err)
				}

			case client.ResolveStrategyInteractive:
				if err := client.ClientSyncInteractive(ctx, clientReader, serverGetter, serverSaver, cryptorInst, token, os.Stdin); err != nil {
					return fmt.Errorf("interactive sync failed: %w", err)
				}

			default:
				return fmt.Errorf("unknown sync mode: %s", syncMode)
			}

		case scheme.GRPC:
			grpcConn, err := grpc.New(serverURL, grpc.WithRetryPolicy(grpc.RetryPolicy{
				Count:   3,
				Wait:    1 * time.Second,
				MaxWait: 5 * time.Second,
			}))
			if err != nil {
				return err
			}
			defer grpcConn.Close()

			serverGetter := facades.NewSecretReaderGRPC(grpcConn)
			serverSaver := facades.NewSecretWriterGRPC(grpcConn)

			switch syncMode {
			case client.ResolveStrategyServer:
				return nil

			case client.ResolveStrategyClient:
				if err := client.ClientSyncClient(ctx, clientReader, serverGetter, serverSaver, token); err != nil {
					return fmt.Errorf("client sync failed: %w", err)
				}

			case client.ResolveStrategyInteractive:
				if err := client.ClientSyncInteractive(ctx, clientReader, serverGetter, serverSaver, cryptorInst, token, os.Stdin); err != nil {
					return fmt.Errorf("interactive sync failed: %w", err)
				}

			default:
				return fmt.Errorf("unknown sync mode: %s", syncMode)
			}

		default:
			return errors.New("unsupported scheme")
		}

		return nil

	case client.CommandVersion:
		fmt.Printf("Version: %s\nBuild Date: %s\n", buildVersion, buildDate)
		return nil

	case client.CommandHelp:
		fmt.Println(client.GetHelp())
		return nil

	default:
		return errors.New("unknown command")
	}
}
