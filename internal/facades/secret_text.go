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

// SecretTextListFacade предоставляет методы для получения списка текстовых секретов через HTTP.
type SecretTextListFacade struct {
	client *resty.Client
}

// NewTextListFacade создает новый экземпляр SecretTextListFacade.
func NewTextListFacade(client *resty.Client) *SecretTextListFacade {
	return &SecretTextListFacade{client: client}
}

// List возвращает список текстовых секретов, полученных через HTTP API.
// Принимает контекст для управления запросом.
// Возвращает срез моделей SecretTextClient или ошибку.
func (f *SecretTextListFacade) List(ctx context.Context) ([]models.SecretTextClient, error) {
	var secrets []models.SecretTextClient

	resp, err := f.client.R().
		SetContext(ctx).
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

// SecretTextListGRPCFacade предоставляет методы для получения списка текстовых секретов через gRPC.
type SecretTextListGRPCFacade struct {
	client pb.SecretTextServiceClient
}

// NewTextListGRPCFacade создает новый экземпляр SecretTextListGRPCFacade.
func NewTextListGRPCFacade(client pb.SecretTextServiceClient) *SecretTextListGRPCFacade {
	return &SecretTextListGRPCFacade{client: client}
}

// List возвращает список текстовых секретов, полученных через gRPC API.
// Принимает контекст и токен аутентификации.
// Возвращает срез моделей SecretTextClient или ошибку.
func (f *SecretTextListGRPCFacade) List(ctx context.Context, token string) ([]models.SecretTextClient, error) {
	// Добавляем токен в metadata для аутентификации
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := f.client.ListTextSecrets(ctx, &pb.SecretTextListRequest{Token: token})
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
