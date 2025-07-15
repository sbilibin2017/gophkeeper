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

// SecretTextSaveHTTP сохраняет текстовый секрет через HTTP.
func SaveSecretTextHTTP(ctx context.Context, client *resty.Client, token string, secret models.SecretTextSaveRequest) error {
	reqBody := map[string]interface{}{
		"secret_name": secret.SecretName,
		"content":     secret.Content,
	}
	if secret.Meta != nil {
		reqBody["meta"] = *secret.Meta
	}

	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetBody(reqBody).
		Post("/save/secret-text")
	if err != nil {
		return fmt.Errorf("failed to send save request: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("server error: %s", resp.Status())
	}
	return nil
}

// SecretTextGetHTTP получает текстовый секрет по имени через HTTP.
func GetSecretTextHTTP(ctx context.Context, client *resty.Client, token, secretName string) (*models.SecretTextGetResponse, error) {
	var secret models.SecretTextGetResponse
	url := fmt.Sprintf("/get/secret-text/%s", secretName)
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error fetching text secret: %s", resp.Status())
	}
	return &secret, nil
}

// SecretTextListHTTP получает список всех текстовых секретов через HTTP.
func ListSecretTextHTTP(ctx context.Context, client *resty.Client, token string) ([]models.SecretTextGetResponse, error) {
	var secrets []models.SecretTextGetResponse
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secrets).
		Get("/list/secret-text")
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error fetching text secrets: %s", resp.Status())
	}
	return secrets, nil
}

// SecretTextSaveGRPC сохраняет текстовый секрет через gRPC.
func SaveSecretTextGRPC(ctx context.Context, client pb.SecretTextServiceClient, token string, secret models.SecretTextSaveRequest) error {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	var meta string
	if secret.Meta != nil {
		meta = *secret.Meta
	} else {
		meta = ""
	}

	req := &pb.SecretTextSaveRequest{
		SecretName: secret.SecretName,
		Content:    secret.Content,
		Meta:       meta,
	}

	_, err := client.Save(ctx, req)
	return err
}

// SecretTextGetGRPC получает текстовый секрет по имени через gRPC.
func GetSecretTextGRPC(ctx context.Context, client pb.SecretTextServiceClient, token, secretName string) (*models.SecretTextGetResponse, error) {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &pb.SecretTextGetRequest{SecretName: secretName}
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

	return &models.SecretTextGetResponse{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Content:     resp.Content,
		Meta:        meta,
		UpdatedAt:   updatedAt,
	}, nil
}

// SecretTextListGRPC получает список всех текстовых секретов через gRPC.
func ListSecretTextGRPC(ctx context.Context, client pb.SecretTextServiceClient, token string) ([]models.SecretTextGetResponse, error) {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctx, &pb.SecretTextListRequest{})
	if err != nil {
		return nil, err
	}

	var result []models.SecretTextGetResponse
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

		result = append(result, models.SecretTextGetResponse{
			SecretName:  item.SecretName,
			SecretOwner: item.SecretOwner,
			Content:     item.Content,
			Meta:        meta,
			UpdatedAt:   updatedAt,
		})
	}

	return result, nil
}

// SaveSecretTextRequest сохраняет или обновляет текстовый секрет (без updated_at и secret_owner).
func SaveSecretTextRequest(ctx context.Context, db *sqlx.DB, secret models.SecretTextSaveRequest) error {
	query := `
		INSERT INTO secret_text_request (secret_name, content, meta)
		VALUES (:secret_name, :content, :meta)
		ON CONFLICT(secret_name) DO UPDATE SET
			content = excluded.content,
			meta = excluded.meta;
	`

	_, err := db.NamedExecContext(ctx, query, secret)
	return err
}

// GetAllSecretsTextRequest возвращает список всех текстовых секретов (только имена).
func GetAllSecretsTextRequest(ctx context.Context, db *sqlx.DB) ([]models.SecretTextGetRequest, error) {
	query := `
		SELECT secret_name
		FROM secret_text_request;
	`

	var secrets []models.SecretTextGetRequest
	if err := db.SelectContext(ctx, &secrets, query); err != nil {
		return nil, err
	}

	return secrets, nil
}

// GetSecretTextByNameRequest возвращает текстовый секрет по secret_name (без owner).
func GetSecretTextByNameRequest(ctx context.Context, db *sqlx.DB, secretName string) (*models.SecretTextGetResponse, error) {
	query := `
		SELECT secret_name, content, meta, updated_at
		FROM secret_text_request
		WHERE secret_name = $1;
	`

	var secret models.SecretTextGetResponse
	err := db.GetContext(ctx, &secret, query, secretName)
	if err != nil {
		return nil, errors.New("secret not found or error fetching")
	}

	return &secret, nil
}
