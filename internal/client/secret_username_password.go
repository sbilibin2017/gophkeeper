package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"google.golang.org/grpc/metadata"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- HTTP ---

// SecretUsernamePasswordSaveHTTP сохраняет секрет с логином и паролем через HTTP.
func SaveSecretUsernamePasswordHTTP(ctx context.Context, client *resty.Client, token string, secret models.SecretUsernamePasswordSaveRequest) error {
	reqBody := map[string]interface{}{
		"secret_name": secret.SecretName,
		"username":    secret.Username,
		"password":    secret.Password,
	}
	if secret.Meta != nil {
		reqBody["meta"] = *secret.Meta
	}

	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetBody(reqBody).
		Post("/save/secret-username-password")
	if err != nil {
		return fmt.Errorf("failed to send save request: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("server error: %s", resp.Status())
	}
	return nil
}

// GetSecretUsernamePasswordHTTP получает секрет с логином и паролем по имени через HTTP.
func GetSecretUsernamePasswordHTTP(ctx context.Context, client *resty.Client, token, secretName string) (*models.SecretUsernamePasswordGetResponse, error) {
	var secret models.SecretUsernamePasswordGetResponse
	url := fmt.Sprintf("/get/secret-username-password/%s", secretName)
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error fetching username-password secret: %s", resp.Status())
	}
	return &secret, nil
}

// SecretUsernamePasswordListHTTP получает список всех секретов с логином и паролем через HTTP.
func ListSecretUsernamePasswordHTTP(ctx context.Context, client *resty.Client, token string) ([]models.SecretUsernamePasswordGetResponse, error) {
	var secrets []models.SecretUsernamePasswordGetResponse
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secrets).
		Get("/list/secret-username-password")
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error fetching username-password secrets: %s", resp.Status())
	}
	return secrets, nil
}

// --- gRPC ---

// SecretUsernamePasswordSaveGRPC сохраняет секрет с логином и паролем через gRPC.
func SaveSecretUsernamePasswordGRPC(ctx context.Context, client pb.SecretUsernamePasswordServiceClient, token string, secret models.SecretUsernamePasswordSaveRequest) error {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	var meta string
	if secret.Meta != nil {
		meta = *secret.Meta
	} else {
		meta = ""
	}

	req := &pb.SecretUsernamePasswordSaveRequest{
		SecretName: secret.SecretName,
		Username:   secret.Username,
		Password:   secret.Password,
		Meta:       meta,
	}

	_, err := client.Save(ctx, req)
	return err
}

// SecretUsernamePasswordGetGRPC получает секрет с логином и паролем по имени через gRPC.
func GetSecretUsernamePasswordGRPC(ctx context.Context, client pb.SecretUsernamePasswordServiceClient, token, secretName string) (*models.SecretUsernamePasswordGetResponse, error) {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &pb.SecretUsernamePasswordGetRequest{SecretName: secretName}
	resp, err := client.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	var meta *string
	if resp.Meta != "" {
		meta = &resp.Meta
	}

	var updatedAt *time.Time
	if resp.UpdatedAt != "" {
		t, err := time.Parse(time.RFC3339, resp.UpdatedAt)
		if err == nil {
			updatedAt = &t
		}
	}

	return &models.SecretUsernamePasswordGetResponse{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Username:    resp.Username,
		Password:    resp.Password,
		Meta:        meta,
		UpdatedAt:   updatedAt,
	}, nil
}

// SecretUsernamePasswordListGRPC получает список всех секретов с логином и паролем через gRPC.
func ListSecretUsernamePasswordGRPC(ctx context.Context, client pb.SecretUsernamePasswordServiceClient, token string) ([]models.SecretUsernamePasswordGetResponse, error) {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctx, &pb.SecretUsernamePasswordListRequest{})
	if err != nil {
		return nil, err
	}

	var result []models.SecretUsernamePasswordGetResponse
	for _, item := range resp.Items {
		var meta *string
		if item.Meta != "" {
			meta = &item.Meta
		}

		var updatedAt *time.Time
		if item.UpdatedAt != "" {
			t, err := time.Parse(time.RFC3339, item.UpdatedAt)
			if err == nil {
				updatedAt = &t
			}
		}

		result = append(result, models.SecretUsernamePasswordGetResponse{
			SecretName:  item.SecretName,
			SecretOwner: item.SecretOwner,
			Username:    item.Username,
			Password:    item.Password,
			Meta:        meta,
			UpdatedAt:   updatedAt,
		})
	}

	return result, nil
}

// SaveSecretUsernamePasswordRequest сохраняет или обновляет секрет с логином и паролем без поля secret_owner и updated_at.
func SaveSecretUsernamePasswordRequest(ctx context.Context, db *sqlx.DB, secret models.SecretUsernamePasswordSaveRequest) error {
	query := `
		INSERT INTO secret_username_password_request (secret_name, username, password, meta)
		VALUES (:secret_name, :username, :password, :meta)
		ON CONFLICT(secret_name) DO UPDATE SET
			username = excluded.username,
			password = excluded.password,
			meta = excluded.meta;
	`

	_, err := db.NamedExecContext(ctx, query, secret)
	return err
}

// GetAllSecretsUsernamePasswordRequest возвращает список всех секретов с логином и паролем (только имена).
func GetAllSecretsUsernamePasswordRequest(ctx context.Context, db *sqlx.DB) ([]models.SecretUsernamePasswordGetRequest, error) {
	query := `
		SELECT secret_name
		FROM secret_username_password_request;
	`

	var secrets []models.SecretUsernamePasswordGetRequest
	if err := db.SelectContext(ctx, &secrets, query); err != nil {
		return nil, err
	}

	return secrets, nil
}

// GetSecretUsernamePasswordByNameRequest возвращает один секрет по secret_name (без owner).
func GetSecretUsernamePasswordByNameRequest(ctx context.Context, db *sqlx.DB, secretName string) (*models.SecretUsernamePasswordGetResponse, error) {
	query := `
		SELECT secret_name, username, password, meta, updated_at
		FROM secret_username_password_request
		WHERE secret_name = $1;
	`

	var secret models.SecretUsernamePasswordGetResponse
	err := db.GetContext(ctx, &secret, query, secretName)
	if err != nil {
		return nil, errors.New("secret not found or error fetching")
	}

	return &secret, nil
}
