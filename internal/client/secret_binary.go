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

// SecretBinarySaveHTTP сохраняет бинарный секрет через HTTP.
func SaveSecretBinaryHTTP(ctx context.Context, client *resty.Client, token string, secret models.SecretBinarySaveRequest) error {
	reqBody := map[string]interface{}{
		"secret_name": secret.SecretName,
		"data":        secret.Data,
	}
	if secret.Meta != nil {
		reqBody["meta"] = *secret.Meta
	}

	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetBody(reqBody).
		Post("/save/secret-binary")
	if err != nil {
		return fmt.Errorf("failed to send save request: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("server error: %s", resp.Status())
	}
	return nil
}

// SecretBinaryGetHTTP получает бинарный секрет по имени через HTTP.
func GetSecretBinaryHTTP(ctx context.Context, client *resty.Client, token, secretName string) (*models.SecretBinaryGetResponse, error) {
	var secret models.SecretBinaryGetResponse
	url := fmt.Sprintf("/get/secret-binary/%s", secretName)
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error fetching binary secret: %s", resp.Status())
	}
	return &secret, nil
}

// SecretBinaryListHTTP получает список всех бинарных секретов через HTTP.
func ListSecretBinaryHTTP(ctx context.Context, client *resty.Client, token string) ([]models.SecretBinaryGetResponse, error) {
	var secrets []models.SecretBinaryGetResponse
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secrets).
		Get("/list/secret-binary")
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error fetching binary secrets: %s", resp.Status())
	}
	return secrets, nil
}

// SecretBinarySaveGRPC сохраняет бинарный секрет через gRPC.
func SaveSecretBinaryGRPC(ctx context.Context, client pb.SecretBinaryServiceClient, token string, secret models.SecretBinarySaveRequest) error {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	var meta string
	if secret.Meta != nil {
		meta = *secret.Meta
	} else {
		meta = ""
	}

	req := &pb.SecretBinarySaveRequest{
		SecretName: secret.SecretName,
		Data:       secret.Data,
		Meta:       meta,
	}

	_, err := client.Save(ctx, req)
	return err
}

// SecretBinaryGetGRPC получает бинарный секрет по имени через gRPC.
func GetSecretBinaryGRPC(ctx context.Context, client pb.SecretBinaryServiceClient, token, secretName string) (*models.SecretBinaryGetResponse, error) {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &pb.SecretBinaryGetRequest{SecretName: secretName}
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

	return &models.SecretBinaryGetResponse{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Data:        resp.Data,
		Meta:        meta,
		UpdatedAt:   updatedAt,
	}, nil
}

// SecretBinaryListGRPC получает список всех бинарных секретов через gRPC.
func ListSecretBinaryGRPC(ctx context.Context, client pb.SecretBinaryServiceClient, token string) ([]models.SecretBinaryGetResponse, error) {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctx, &pb.SecretBinaryListRequest{})
	if err != nil {
		return nil, err
	}

	var result []models.SecretBinaryGetResponse
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

		result = append(result, models.SecretBinaryGetResponse{
			SecretName:  item.SecretName,
			SecretOwner: item.SecretOwner,
			Data:        item.Data,
			Meta:        meta,
			UpdatedAt:   updatedAt,
		})
	}

	return result, nil
}

// SaveSecretBinaryRequest сохраняет или обновляет бинарный секрет (без updated_at и secret_owner).
func SaveSecretBinaryRequest(ctx context.Context, db *sqlx.DB, secret models.SecretBinarySaveRequest) error {
	query := `
		INSERT INTO secret_binary_request (secret_name, data, meta)
		VALUES (:secret_name, :data, :meta)
		ON CONFLICT(secret_name) DO UPDATE SET
			data = excluded.data,
			meta = excluded.meta;
	`

	_, err := db.NamedExecContext(ctx, query, secret)
	return err
}

// GetAllSecretsBinaryRequest возвращает список всех бинарных секретов (только имена).
func GetAllSecretsBinaryRequest(ctx context.Context, db *sqlx.DB) ([]models.SecretBinaryGetRequest, error) {
	query := `
		SELECT secret_name
		FROM secret_binary_request;
	`

	var secrets []models.SecretBinaryGetRequest
	if err := db.SelectContext(ctx, &secrets, query); err != nil {
		return nil, err
	}

	return secrets, nil
}

// GetSecretBinaryByNameRequest возвращает бинарный секрет по secret_name (без owner).
func GetSecretBinaryByNameRequest(ctx context.Context, db *sqlx.DB, secretName string) (*models.SecretBinaryGetResponse, error) {
	query := `
		SELECT secret_name, data, meta, updated_at
		FROM secret_binary_request
		WHERE secret_name = $1;
	`

	var secret models.SecretBinaryGetResponse
	err := db.GetContext(ctx, &secret, query, secretName)
	if err != nil {
		return nil, errors.New("secret not found or error fetching")
	}

	return &secret, nil
}
