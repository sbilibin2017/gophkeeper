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

// SecretBinaryListFacade предоставляет методы для получения списка бинарных секретов через HTTP.
type SecretBinaryListFacade struct {
	client *resty.Client
}

// NewBinaryListFacade создает новый экземпляр SecretBinaryListFacade.
func NewBinaryListFacade(client *resty.Client) *SecretBinaryListFacade {
	return &SecretBinaryListFacade{client: client}
}

// List возвращает список бинарных секретов, полученных через HTTP API.
// Принимает контекст для управления запросом.
// Возвращает срез моделей SecretBinaryClient или ошибку.
func (f *SecretBinaryListFacade) List(ctx context.Context) ([]models.SecretBinaryClient, error) {
	var secrets []models.SecretBinaryClient

	resp, err := f.client.R().
		SetContext(ctx).
		SetResult(&secrets).
		Get("/list/secret-binary")
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, errors.New("error fetching binary secrets: " + resp.Status())
	}
	return secrets, nil
}

// SecretBinaryListGRPCFacade предоставляет методы для получения списка бинарных секретов через gRPC.
type SecretBinaryListGRPCFacade struct {
	client pb.SecretBinaryServiceClient
}

// NewBinaryListGRPCFacade создает новый экземпляр SecretBinaryListGRPCFacade.
func NewBinaryListGRPCFacade(client pb.SecretBinaryServiceClient) *SecretBinaryListGRPCFacade {
	return &SecretBinaryListGRPCFacade{client: client}
}

// List возвращает список бинарных секретов, полученных через gRPC API.
// Принимает контекст и токен аутентификации.
// Возвращает срез моделей SecretBinaryClient или ошибку.
func (f *SecretBinaryListGRPCFacade) List(ctx context.Context, token string) ([]models.SecretBinaryClient, error) {
	// Добавляем токен в metadata для аутентификации
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Выполняем gRPC запрос
	resp, err := f.client.ListBinarySecrets(ctx, &pb.SecretBinaryListRequest{Token: token})
	if err != nil {
		return nil, err
	}

	// Преобразуем protobuf-объекты в внутренние модели
	var result []models.SecretBinaryClient
	for _, item := range resp.Items {
		updatedAt, err := time.Parse(time.RFC3339, item.UpdatedAt)
		if err != nil {
			return nil, errors.New("invalid updated_at format in response")
		}

		var meta *string
		if item.Meta != "" {
			meta = &item.Meta
		}

		result = append(result, models.SecretBinaryClient{
			SecretName: item.SecretName,
			Data:       item.Data,
			Meta:       meta,
			UpdatedAt:  updatedAt,
		})
	}

	return result, nil
}
