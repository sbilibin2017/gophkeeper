package facades

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"google.golang.org/grpc/metadata"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// SecretTextListFacade - HTTP фасад (оставляем токен в SetAuthToken)
type SecretTextListFacade struct {
	client *resty.Client
}

func NewTextListFacade(client *resty.Client) *SecretTextListFacade {
	return &SecretTextListFacade{client: client}
}

func (f *SecretTextListFacade) List(ctx context.Context, token string) ([]models.SecretTextClient, error) {
	var secrets []models.SecretTextClient

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secrets).
		Get("/list/secret-text")
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, errors.New("error fetching text secrets: " + resp.Status())
	}
	return secrets, nil
}

// SecretTextListGRPCFacade - gRPC фасад
type SecretTextListGRPCFacade struct {
	client pb.SecretTextServiceClient
}

func NewTextListGRPCFacade(client pb.SecretTextServiceClient) *SecretTextListGRPCFacade {
	return &SecretTextListGRPCFacade{client: client}
}

func (f *SecretTextListGRPCFacade) List(ctx context.Context, token string) ([]models.SecretTextClient, error) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Убираем поле Token из запроса, он теперь только в metadata
	resp, err := f.client.List(ctx, &pb.SecretTextListRequest{})
	if err != nil {
		return nil, err
	}

	var result []models.SecretTextClient
	for _, item := range resp.Items {
		updatedAt, err := time.Parse(time.RFC3339, item.UpdatedAt)
		if err != nil {
			return nil, errors.New("invalid updated_at format in response")
		}

		var meta *string
		if item.Meta != "" {
			meta = &item.Meta
		}

		result = append(result, models.SecretTextClient{
			SecretName: item.SecretName,
			Content:    item.Content,
			Meta:       meta,
			UpdatedAt:  updatedAt,
		})
	}

	return result, nil
}

// SecretTextSaveHTTPFacade - HTTP фасад
type SecretTextSaveHTTPFacade struct {
	client *resty.Client
}

func NewSecretTextSaveHTTPFacade(client *resty.Client) *SecretTextSaveHTTPFacade {
	return &SecretTextSaveHTTPFacade{client: client}
}

func (f *SecretTextSaveHTTPFacade) Save(ctx context.Context, token string, secret models.SecretTextClient) error {
	reqBody := map[string]interface{}{
		"secret_name": secret.SecretName,
		"content":     secret.Content,
		"updated_at":  secret.UpdatedAt.Format(time.RFC3339),
	}
	if secret.Meta != nil {
		reqBody["meta"] = *secret.Meta
	}

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetBody(reqBody).
		Post("/save/secret-text")

	if err != nil {
		return fmt.Errorf("failed to send save request: %w", err)
	}

	if resp.IsError() {
		return errors.New("server error: " + resp.Status())
	}

	return nil
}

// SecretTextGetHTTPFacade - HTTP фасад
type SecretTextGetHTTPFacade struct {
	client *resty.Client
}

func NewSecretTextGetHTTPFacade(client *resty.Client) *SecretTextGetHTTPFacade {
	return &SecretTextGetHTTPFacade{client: client}
}

func (f *SecretTextGetHTTPFacade) Get(ctx context.Context, token string, secretName string) (*models.SecretTextClient, error) {
	var secret models.SecretTextClient

	url := fmt.Sprintf("/get/secret-text/%s", secretName)

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, errors.New("error fetching text secret: " + resp.Status())
	}

	return &secret, nil
}

// SecretTextGetGRPCFacade - gRPC фасад
type SecretTextGetGRPCFacade struct {
	client pb.SecretTextServiceClient
}

func NewSecretTextGetGRPCFacade(client pb.SecretTextServiceClient) *SecretTextGetGRPCFacade {
	return &SecretTextGetGRPCFacade{client: client}
}

func (f *SecretTextGetGRPCFacade) Get(ctx context.Context, token string, secretName string) (*models.SecretTextClient, error) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Убираем token из запроса
	req := &pb.SecretTextGetRequest{
		SecretName: secretName,
	}

	resp, err := f.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	item := resp.Secret
	updatedAt, err := time.Parse(time.RFC3339, item.UpdatedAt)
	if err != nil {
		return nil, errors.New("invalid updated_at format in response")
	}

	var meta *string
	if item.Meta != "" {
		meta = &item.Meta
	}

	return &models.SecretTextClient{
		SecretName: item.SecretName,
		Content:    item.Content,
		Meta:       meta,
		UpdatedAt:  updatedAt,
	}, nil
}

// SecretTextSaveGRPCFacade - gRPC фасад
type SecretTextSaveGRPCFacade struct {
	client pb.SecretTextServiceClient
}

func NewSecretTextSaveGRPCFacade(client pb.SecretTextServiceClient) *SecretTextSaveGRPCFacade {
	return &SecretTextSaveGRPCFacade{client: client}
}

func (f *SecretTextSaveGRPCFacade) Save(ctx context.Context, token string, secret models.SecretTextClient) error {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	var meta string
	if secret.Meta != nil {
		meta = *secret.Meta
	}

	req := &pb.SecretTextSaveRequest{
		Secret: &pb.SecretText{
			SecretName: secret.SecretName,
			Content:    secret.Content,
			Meta:       meta,
			UpdatedAt:  secret.UpdatedAt.Format(time.RFC3339),
		},
	}

	_, err := f.client.Save(ctx, req)
	return err
}
