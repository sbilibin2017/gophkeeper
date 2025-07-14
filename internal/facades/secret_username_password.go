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

// SecretUsernamePasswordListFacade предоставляет методы для получения списка секретов "логин-пароль" через HTTP.
type SecretUsernamePasswordListFacade struct {
	client *resty.Client
}

// NewUsernamePasswordListFacade создает новый экземпляр SecretUsernamePasswordListFacade.
func NewUsernamePasswordListFacade(client *resty.Client) *SecretUsernamePasswordListFacade {
	return &SecretUsernamePasswordListFacade{client: client}
}

// List возвращает список секретов "логин-пароль", полученных через HTTP API.
// Принимает контекст для управления временем выполнения запроса.
// Возвращает срез моделей SecretUsernamePasswordClient или ошибку.
func (f *SecretUsernamePasswordListFacade) List(ctx context.Context) ([]models.SecretUsernamePasswordClient, error) {
	var secrets []models.SecretUsernamePasswordClient

	resp, err := f.client.R().
		SetContext(ctx).
		SetResult(&secrets).
		Get("/list/secret-username-password")
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, errors.New("error fetching username-password secrets: " + resp.Status())
	}
	return secrets, nil
}

// SecretUsernamePasswordListGRPCFacade предоставляет методы для получения списка секретов "логин-пароль" через gRPC.
type SecretUsernamePasswordListGRPCFacade struct {
	client pb.SecretUsernamePasswordServiceClient
}

// NewUsernamePasswordListGRPCFacade создает новый экземпляр SecretUsernamePasswordListGRPCFacade.
func NewUsernamePasswordListGRPCFacade(client pb.SecretUsernamePasswordServiceClient) *SecretUsernamePasswordListGRPCFacade {
	return &SecretUsernamePasswordListGRPCFacade{client: client}
}

// List возвращает список секретов "логин-пароль", полученных через gRPC API.
// Принимает контекст и токен аутентификации.
// Возвращает срез моделей SecretUsernamePasswordClient или ошибку.
func (f *SecretUsernamePasswordListGRPCFacade) List(ctx context.Context, token string) ([]models.SecretUsernamePasswordClient, error) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := f.client.ListUsernamePasswordSecrets(ctx, &pb.SecretUsernamePasswordListRequest{Token: token})
	if err != nil {
		return nil, err
	}

	var result []models.SecretUsernamePasswordClient
	for _, item := range resp.Items {
		updatedAt, err := time.Parse(time.RFC3339, item.UpdatedAt)
		if err != nil {
			return nil, errors.New("invalid updated_at format in response")
		}

		var meta *string
		if item.Meta != "" {
			meta = &item.Meta
		}

		result = append(result, models.SecretUsernamePasswordClient{
			SecretName: item.SecretName,
			Username:   item.Username,
			Password:   item.Password,
			Meta:       meta,
			UpdatedAt:  updatedAt,
		})
	}

	return result, nil
}
