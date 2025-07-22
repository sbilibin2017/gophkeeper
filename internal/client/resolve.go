package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func ResolveBankCardHTTP(
	ctx context.Context,
	reader *bufio.Reader,
	strategy string,
	listClientFunc func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error),
	getServerFunc func(ctx context.Context, client *resty.Client, secretName string) (*models.BankCardDB, error),
	addServerFunc func(ctx context.Context, client *resty.Client, req *models.BankCardAddRequest) error,
	db *sqlx.DB,
	client *resty.Client,
	secretName string,
) error {
	toBankCardAddRequest := func(secret models.BankCardDB) *models.BankCardAddRequest {
		return &models.BankCardAddRequest{
			SecretName: secret.SecretName,
			Number:     secret.Number,
			Owner:      secret.Owner,
			Exp:        secret.Exp,
			CVV:        secret.CVV,
			Meta:       secret.Meta,
		}
	}

	switch strategy {
	case "server":
		// No action needed for "server" strategy here
		return nil

	case "client":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}

		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			if secretServer == nil || secretServer.UpdatedAt.Before(secretClient.UpdatedAt) {
				if err := addServerFunc(ctx, client, toBankCardAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			}
		}
		return nil

	case "interactive":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}

		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			if secretServer == nil {
				if err := addServerFunc(ctx, client, toBankCardAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
				continue
			}

			if !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			// Conflict detected
			fmt.Printf("Conflict detected for secret '%s':\n", secretClient.SecretName)
			fmt.Printf("Server version: %+v\n", secretServer)
			fmt.Printf("Client version: %+v\n", secretClient)
			fmt.Print("Choose version to keep (server/client): ")

			choiceRaw, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			choice := strings.TrimSpace(choiceRaw)

			switch choice {
			case "server":
				continue
			case "client":
				if err := addServerFunc(ctx, client, toBankCardAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			default:
				return errors.New("invalid choice, expected 'server' or 'client'")
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func ResolveBankCardGRPC(
	ctx context.Context,
	reader *bufio.Reader,
	strategy string,
	listClientFunc func(ctx context.Context, db *sqlx.DB) ([]models.BankCardDB, error),
	getServerFunc func(ctx context.Context, client pb.BankCardServiceClient, secretName string) (*models.BankCardDB, error),
	addServerFunc func(ctx context.Context, client pb.BankCardServiceClient, req *models.BankCardAddRequest) error,
	db *sqlx.DB,
	client pb.BankCardServiceClient,
	secretName string,
) error {
	// Conversion helper: BankCardDB → BankCardAddRequest
	toBankCardAddRequest := func(secret models.BankCardDB) *models.BankCardAddRequest {
		return &models.BankCardAddRequest{
			SecretName: secret.SecretName,
			Number:     secret.Number,
			Owner:      secret.Owner,
			Exp:        secret.Exp,
			CVV:        secret.CVV,
			Meta:       secret.Meta,
		}
	}

	switch strategy {
	case "server":
		return nil

	case "client":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}

		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			if secretServer == nil || secretServer.UpdatedAt.Before(secretClient.UpdatedAt) {
				if err := addServerFunc(ctx, client, toBankCardAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			}
		}
		return nil

	case "interactive":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}

		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			if secretServer == nil {
				// Server has no secret, add client secret
				if err := addServerFunc(ctx, client, toBankCardAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
				continue
			}

			if !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			// Conflict detected
			fmt.Printf("Conflict detected for secret '%s':\n", secretClient.SecretName)
			fmt.Printf("Server version: %+v\n", secretServer)
			fmt.Printf("Client version: %+v\n", secretClient)
			fmt.Print("Choose version to keep (server/client): ")

			choiceRaw, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			choice := strings.TrimSpace(choiceRaw)

			switch choice {
			case "server":
				// Keep server version — do nothing
			case "client":
				if err := addServerFunc(ctx, client, toBankCardAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			default:
				return errors.New("invalid choice, expected 'server' or 'client'")
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func ResolveTextHTTP(
	ctx context.Context,
	reader *bufio.Reader,
	strategy string,
	listClientFunc func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error),
	getServerFunc func(ctx context.Context, client *resty.Client, secretName string) (*models.TextDB, error),
	addServerFunc func(ctx context.Context, client *resty.Client, req *models.TextAddRequest) error,
	db *sqlx.DB,
	client *resty.Client,
	secretName string,
) error {
	toTextAddRequest := func(secret models.TextDB) *models.TextAddRequest {
		return &models.TextAddRequest{
			SecretName: secret.SecretName,
			Content:    secret.Content,
			Meta:       secret.Meta,
		}
	}

	switch strategy {
	case "server":
		return nil

	case "client":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			if secretServer == nil || secretServer.UpdatedAt.Before(secretClient.UpdatedAt) {
				if err := addServerFunc(ctx, client, toTextAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			}
		}
		return nil

	case "interactive":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			if secretServer == nil {
				if err := addServerFunc(ctx, client, toTextAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
				continue
			}

			if !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			fmt.Printf("Conflict detected for secret '%s':\n", secretClient.SecretName)
			fmt.Printf("Server version: %+v\n", secretServer)
			fmt.Printf("Client version: %+v\n", secretClient)
			fmt.Print("Choose version to keep (server/client): ")

			choiceRaw, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			choice := strings.TrimSpace(choiceRaw)

			switch choice {
			case "server":
				// Keep server version — do nothing
			case "client":
				if err := addServerFunc(ctx, client, toTextAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			default:
				return errors.New("invalid choice, expected 'server' or 'client'")
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func ResolveTextGRPC(
	ctx context.Context,
	reader *bufio.Reader,
	strategy string,
	listClientFunc func(ctx context.Context, db *sqlx.DB) ([]models.TextDB, error),
	getServerFunc func(ctx context.Context, client pb.TextServiceClient, secretName string) (*models.TextDB, error),
	addServerFunc func(ctx context.Context, client pb.TextServiceClient, req *models.TextAddRequest) error,
	db *sqlx.DB,
	client pb.TextServiceClient,
	secretName string,
) error {
	toTextAddRequest := func(secret models.TextDB) *models.TextAddRequest {
		return &models.TextAddRequest{
			SecretName: secret.SecretName,
			Content:    secret.Content,
			Meta:       secret.Meta,
		}
	}

	switch strategy {
	case "server":
		return nil

	case "client":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			// If server secret exists and is newer or same timestamp, skip syncing client secret
			if secretServer != nil && !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			if err := addServerFunc(ctx, client, toTextAddRequest(secretClient)); err != nil {
				return fmt.Errorf("failed to add client secret to server: %w", err)
			}
		}
		return nil

	case "interactive":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}
			if secretServer == nil {
				if err := addServerFunc(ctx, client, toTextAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
				continue
			}
			if !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			fmt.Printf("Conflict detected for secret '%s':\n", secretClient.SecretName)
			fmt.Printf("Server version: %+v\n", secretServer)
			fmt.Printf("Client version: %+v\n", secretClient)
			fmt.Print("Choose version to keep (server/client): ")

			choiceRaw, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			choice := strings.TrimSpace(choiceRaw)

			switch choice {
			case "server":
				// Keep server version
			case "client":
				if err := addServerFunc(ctx, client, toTextAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			default:
				return errors.New("invalid choice, expected 'server' or 'client'")
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func ResolveBinaryHTTP(
	ctx context.Context,
	reader *bufio.Reader,
	strategy string,
	listClientFunc func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error),
	getServerFunc func(ctx context.Context, client *resty.Client, secretName string) (*models.BinaryDB, error),
	addServerFunc func(ctx context.Context, client *resty.Client, req *models.BinaryAddRequest) error,
	db *sqlx.DB,
	client *resty.Client,
	secretName string,
) error {
	toBinaryAddRequest := func(secret models.BinaryDB) *models.BinaryAddRequest {
		return &models.BinaryAddRequest{
			SecretName: secret.SecretName,
			Data:       secret.Data,
			Meta:       secret.Meta,
		}
	}

	switch strategy {
	case "server":
		return nil

	case "client":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			if err := addServerFunc(ctx, client, toBinaryAddRequest(secretClient)); err != nil {
				return fmt.Errorf("failed to add client secret to server: %w", err)
			}
		}
		return nil

	case "interactive":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}
			if secretServer == nil {
				if err := addServerFunc(ctx, client, toBinaryAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
				continue
			}
			if !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			fmt.Printf("Conflict detected for secret '%s':\n", secretClient.SecretName)
			fmt.Printf("Server version: %+v\n", secretServer)
			fmt.Printf("Client version: %+v\n", secretClient)
			fmt.Print("Choose version to keep (server/client): ")

			choiceRaw, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			choice := strings.TrimSpace(choiceRaw)

			switch choice {
			case "server":
				// Keep server version
			case "client":
				if err := addServerFunc(ctx, client, toBinaryAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			default:
				return errors.New("invalid choice, expected 'server' or 'client'")
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func ResolveBinaryGRPC(
	ctx context.Context,
	reader *bufio.Reader,
	strategy string,
	listClientFunc func(ctx context.Context, db *sqlx.DB) ([]models.BinaryDB, error),
	getServerFunc func(ctx context.Context, client pb.BinaryServiceClient, secretName string) (*models.BinaryDB, error),
	addServerFunc func(ctx context.Context, client pb.BinaryServiceClient, req *models.BinaryAddRequest) error,
	db *sqlx.DB,
	client pb.BinaryServiceClient,
	secretName string,
) error {
	toBinaryAddRequest := func(secret models.BinaryDB) *models.BinaryAddRequest {
		return &models.BinaryAddRequest{
			SecretName: secret.SecretName,
			Data:       secret.Data,
			Meta:       secret.Meta,
		}
	}

	switch strategy {
	case "server":
		return nil

	case "client":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			if secretServer != nil && !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			if err := addServerFunc(ctx, client, toBinaryAddRequest(secretClient)); err != nil {
				return fmt.Errorf("failed to add client secret to server: %w", err)
			}
		}
		return nil

	case "interactive":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			if secretServer == nil {
				if err := addServerFunc(ctx, client, toBinaryAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
				continue
			}

			if !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			fmt.Printf("Conflict detected for secret '%s':\n", secretClient.SecretName)
			fmt.Printf("Server version: %+v\n", secretServer)
			fmt.Printf("Client version: %+v\n", secretClient)
			fmt.Print("Choose version to keep (server/client): ")

			choiceRaw, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			choice := strings.TrimSpace(choiceRaw)

			switch choice {
			case "server":
				// Keep server version
			case "client":
				if err := addServerFunc(ctx, client, toBinaryAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			default:
				return errors.New("invalid choice, expected 'server' or 'client'")
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func ResolveUserHTTP(
	ctx context.Context,
	reader *bufio.Reader,
	strategy string,
	listClientFunc func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error),
	getServerFunc func(ctx context.Context, client *resty.Client, secretName string) (*models.UserDB, error),
	addServerFunc func(ctx context.Context, client *resty.Client, req *models.UserAddRequest) error,
	db *sqlx.DB,
	client *resty.Client,
	secretName string,
) error {
	toUserAddRequest := func(user models.UserDB) *models.UserAddRequest {
		return &models.UserAddRequest{
			SecretName: user.SecretName,
			Username:   user.Username,
			Password:   user.Password, // Ensure this field exists in UserDB
			Meta:       user.Meta,
		}
	}

	switch strategy {
	case "server":
		return nil

	case "client":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			if err := addServerFunc(ctx, client, toUserAddRequest(secretClient)); err != nil {
				return fmt.Errorf("failed to add client secret to server: %w", err)
			}
		}
		return nil

	case "interactive":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			if secretServer == nil {
				if err := addServerFunc(ctx, client, toUserAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
				continue
			}

			if !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			fmt.Printf("Conflict detected for secret '%s':\n", secretClient.SecretName)
			fmt.Printf("Server version: %+v\n", secretServer)
			fmt.Printf("Client version: %+v\n", secretClient)
			fmt.Print("Choose version to keep (server/client): ")

			choiceRaw, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			choice := strings.TrimSpace(choiceRaw)

			switch choice {
			case "server":
				// Keep server version
			case "client":
				if err := addServerFunc(ctx, client, toUserAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			default:
				return errors.New("invalid choice, expected 'server' or 'client'")
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func ResolveUserGRPC(
	ctx context.Context,
	reader *bufio.Reader,
	strategy string,
	listClientFunc func(ctx context.Context, db *sqlx.DB) ([]models.UserDB, error),
	getServerFunc func(ctx context.Context, client pb.UserServiceClient, secretName string) (*models.UserDB, error),
	addServerFunc func(ctx context.Context, client pb.UserServiceClient, req *models.UserAddRequest) error,
	db *sqlx.DB,
	client pb.UserServiceClient,
	secretName string,
) error {
	toUserAddRequest := func(user models.UserDB) *models.UserAddRequest {
		return &models.UserAddRequest{
			SecretName: user.SecretName,
			Username:   user.Username,
			Password:   user.Password, // make sure this field exists in UserDB
			Meta:       user.Meta,
		}
	}

	switch strategy {
	case "server":
		return nil

	case "client":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}

			if secretServer != nil && !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			if err := addServerFunc(ctx, client, toUserAddRequest(secretClient)); err != nil {
				return fmt.Errorf("failed to add client secret to server: %w", err)
			}
		}
		return nil

	case "interactive":
		secrets, err := listClientFunc(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to list client secrets: %w", err)
		}
		for _, secretClient := range secrets {
			secretServer, err := getServerFunc(ctx, client, secretClient.SecretName)
			if err != nil {
				return fmt.Errorf("failed to get secret from server: %w", err)
			}
			if secretServer == nil {
				if err := addServerFunc(ctx, client, toUserAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
				continue
			}
			if !secretClient.UpdatedAt.After(secretServer.UpdatedAt) {
				continue
			}

			fmt.Printf("Conflict detected for secret '%s':\n", secretClient.SecretName)
			fmt.Printf("Server version: %+v\n", secretServer)
			fmt.Printf("Client version: %+v\n", secretClient)
			fmt.Print("Choose version to keep (server/client): ")

			choiceRaw, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			choice := strings.TrimSpace(choiceRaw)

			switch choice {
			case "server":
				// Keep server version
			case "client":
				if err := addServerFunc(ctx, client, toUserAddRequest(secretClient)); err != nil {
					return fmt.Errorf("failed to add client secret to server: %w", err)
				}
			default:
				return errors.New("invalid choice, expected 'server' or 'client'")
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}
