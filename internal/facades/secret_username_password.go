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

// SecretUsernamePasswordSaveHTTPFacade обеспечивает сохранение секретов "логин-пароль" через HTTP.
type SecretUsernamePasswordSaveHTTPFacade struct {
	client *resty.Client
}

// NewSaveUsernamePasswordSecretFacade создаёт новый фасад для сохранения секретов "логин-пароль" через HTTP.
func NewSecretUsernamePasswordSaveHTTPFacade(client *resty.Client) *SecretUsernamePasswordSaveHTTPFacade {
	return &SecretUsernamePasswordSaveHTTPFacade{client: client}
}

// Save отправляет секрет "логин-пароль" на сервер.
// Принимает контекст, токен аутентификации и секрет для сохранения.
// Возвращает ошибку, если операция не удалась.
func (f *SecretUsernamePasswordSaveHTTPFacade) Save(ctx context.Context, token string, secret models.SecretUsernamePasswordClient) error {
	reqBody := map[string]interface{}{
		"secret_name": secret.SecretName,
		"username":    secret.Username,
		"password":    secret.Password,
		"updated_at":  secret.UpdatedAt.Format(time.RFC3339),
	}

	if secret.Meta != nil {
		reqBody["meta"] = *secret.Meta
	}

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetBody(reqBody).
		Post("/save/secret-username-password")

	if err != nil {
		return fmt.Errorf("failed to send save request: %w", err)
	}

	if resp.IsError() {
		return errors.New("server error: " + resp.Status())
	}

	return nil
}

// SecretUsernamePasswordGetHTTPFacade обеспечивает получение одного секрета "логин-пароль" через HTTP.
type SecretUsernamePasswordGetHTTPFacade struct {
	client *resty.Client
}

// NewSecretUsernamePasswordGetHTTPFacade создает новый SecretUsernamePasswordGetHTTPFacade.
func NewSecretUsernamePasswordGetHTTPFacade(client *resty.Client) *SecretUsernamePasswordGetHTTPFacade {
	return &SecretUsernamePasswordGetHTTPFacade{client: client}
}

// Get запрашивает конкретный секрет "логин-пароль" по имени через HTTP API.
// Принимает контекст, токен аутентификации и имя секрета.
// Возвращает указатель на модель SecretUsernamePasswordClient или ошибку.
func (f *SecretUsernamePasswordGetHTTPFacade) Get(ctx context.Context, token string, secretName string) (*models.SecretUsernamePasswordClient, error) {
	var secret models.SecretUsernamePasswordClient

	url := fmt.Sprintf("/get/secret-username-password/%s", secretName)

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, errors.New("error fetching username-password secret: " + resp.Status())
	}

	return &secret, nil
}

// SecretUsernamePasswordListHTTPFacade обеспечивает получение списка секретов "логин-пароль" через HTTP.
type SecretUsernamePasswordListHTTPFacade struct {
	client *resty.Client
}

// NewSecretUsernamePasswordListHTTPFacade создает новый SecretUsernamePasswordListHTTPFacade.
func NewSecretUsernamePasswordListHTTPFacade(client *resty.Client) *SecretUsernamePasswordListHTTPFacade {
	return &SecretUsernamePasswordListHTTPFacade{client: client}
}

// List запрашивает список секретов "логин-пароль" через HTTP API.
func (f *SecretUsernamePasswordListHTTPFacade) List(ctx context.Context, token string) ([]models.SecretUsernamePasswordClient, error) {
	var secrets []models.SecretUsernamePasswordClient

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
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

// --- gRPC фасады ---

// SecretUsernamePasswordSaveGRPCFacade обеспечивает сохранение секретов "логин-пароль" через gRPC.
type SecretUsernamePasswordSaveGRPCFacade struct {
	client pb.SecretUsernamePasswordServiceClient
}

// NewSecretUsernamePasswordSaveGRPCFacade создает новый SecretUsernamePasswordSaveGRPCFacade.
func NewSecretUsernamePasswordSaveGRPCFacade(client pb.SecretUsernamePasswordServiceClient) *SecretUsernamePasswordSaveGRPCFacade {
	return &SecretUsernamePasswordSaveGRPCFacade{client: client}
}

// Save отправляет секрет "логин-пароль" через gRPC.
// Токен передается в metadata, не в теле.
func (f *SecretUsernamePasswordSaveGRPCFacade) Save(ctx context.Context, token string, secret models.SecretUsernamePasswordClient) error {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	var meta string
	if secret.Meta != nil {
		meta = *secret.Meta
	} else {
		meta = ""
	}

	req := &pb.SecretUsernamePasswordSaveRequest{
		Secret: &pb.SecretUsernamePassword{
			SecretName: secret.SecretName,
			Username:   secret.Username,
			Password:   secret.Password,
			Meta:       meta,
			UpdatedAt:  secret.UpdatedAt.Format(time.RFC3339),
		},
	}

	_, err := f.client.Save(ctx, req)
	return err
}

// SecretUsernamePasswordGetGRPCFacade обеспечивает получение одного секрета "логин-пароль" через gRPC.
type SecretUsernamePasswordGetGRPCFacade struct {
	client pb.SecretUsernamePasswordServiceClient
}

// NewSecretUsernamePasswordGetGRPCFacade создает новый SecretUsernamePasswordGetGRPCFacade.
func NewSecretUsernamePasswordGetGRPCFacade(client pb.SecretUsernamePasswordServiceClient) *SecretUsernamePasswordGetGRPCFacade {
	return &SecretUsernamePasswordGetGRPCFacade{client: client}
}

// Get запрашивает секрет "логин-пароль" по имени через gRPC API.
func (f *SecretUsernamePasswordGetGRPCFacade) Get(ctx context.Context, token string, secretName string) (*models.SecretUsernamePasswordClient, error) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &pb.SecretUsernamePasswordGetRequest{
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

	return &models.SecretUsernamePasswordClient{
		SecretName: item.SecretName,
		Username:   item.Username,
		Password:   item.Password,
		Meta:       meta,
		UpdatedAt:  updatedAt,
	}, nil
}

// SecretUsernamePasswordListGRPCFacade обеспечивает получение списка секретов "логин-пароль" через gRPC.
type SecretUsernamePasswordListGRPCFacade struct {
	client pb.SecretUsernamePasswordServiceClient
}

// NewSecretUsernamePasswordListGRPCFacade создает новый SecretUsernamePasswordListGRPCFacade.
func NewSecretUsernamePasswordListGRPCFacade(client pb.SecretUsernamePasswordServiceClient) *SecretUsernamePasswordListGRPCFacade {
	return &SecretUsernamePasswordListGRPCFacade{client: client}
}

// List запрашивает список секретов "логин-пароль" через gRPC API.
func (f *SecretUsernamePasswordListGRPCFacade) List(ctx context.Context, token string) ([]models.SecretUsernamePasswordClient, error) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := f.client.List(ctx, &pb.SecretUsernamePasswordListRequest{})
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
