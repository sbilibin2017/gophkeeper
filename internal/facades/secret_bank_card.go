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

// SecretBankCardSaveHTTPFacade обеспечивает сохранение секретов банковских карт через HTTP.
type SecretBankCardSaveHTTPFacade struct {
	client *resty.Client
}

// NewSaveBankCardSecretFacade создаёт новый фасад для сохранения банковских карт через HTTP.
func NewSecretBankCardSaveHTTPFacade(client *resty.Client) *SecretBankCardSaveHTTPFacade {
	return &SecretBankCardSaveHTTPFacade{client: client}
}

// Save отправляет секрет банковской карты на сервер.
// Принимает контекст, токен аутентификации и секрет для сохранения.
// Возвращает ошибку, если операция не удалась.
func (f *SecretBankCardSaveHTTPFacade) Save(ctx context.Context, token string, secret models.SecretBankCardClient) error {
	reqBody := map[string]interface{}{
		"secret_name": secret.SecretName,
		"owner":       secret.Owner,
		"number":      secret.Number,
		"exp":         secret.Exp,
		"cvv":         secret.CVV,
		"updated_at":  secret.UpdatedAt.Format(time.RFC3339),
	}

	if secret.Meta != nil {
		reqBody["meta"] = *secret.Meta
	}

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetBody(reqBody).
		Post("/save/secret-bank-card")

	if err != nil {
		return fmt.Errorf("failed to send save request: %w", err)
	}

	if resp.IsError() {
		return errors.New("server error: " + resp.Status())
	}

	return nil
}

// SecretBankCardGetHTTPFacade обеспечивает получение одного секрета банковской карты через HTTP.
type SecretBankCardGetHTTPFacade struct {
	client *resty.Client
}

// NewSecretBankCardGetHTTPFacade создает новый SecretBankCardGetHTTPFacade.
func NewSecretBankCardGetHTTPFacade(client *resty.Client) *SecretBankCardGetHTTPFacade {
	return &SecretBankCardGetHTTPFacade{client: client}
}

// Get запрашивает конкретный секрет банковской карты по имени через HTTP API.
// Принимает контекст, токен аутентификации и имя секрета.
// Возвращает указатель на модель SecretBankCardClient или ошибку.
func (f *SecretBankCardGetHTTPFacade) Get(ctx context.Context, token string, secretName string) (*models.SecretBankCardClient, error) {
	var secret models.SecretBankCardClient

	// Формируем путь с именем секрета
	url := fmt.Sprintf("/get/secret-bank-card/%s", secretName)

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, errors.New("error fetching bank card secret: " + resp.Status())
	}

	return &secret, nil
}

type SecretBankCardListHTTPFacade struct {
	client *resty.Client
}

func NewSecretBankCardListHTTPFacade(client *resty.Client) *SecretBankCardListHTTPFacade {
	return &SecretBankCardListHTTPFacade{client: client}
}

func (f *SecretBankCardListHTTPFacade) List(ctx context.Context, token string) ([]models.SecretBankCardClient, error) {
	var secrets []models.SecretBankCardClient

	resp, err := f.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secrets).
		Get("/list/secret-bank-card")
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, errors.New("error fetching bank card secrets: " + resp.Status())
	}
	return secrets, nil
}

// Аналогично для Save и Get HTTP фасадов (оставляем токен в SetAuthToken, т.к. это HTTP, не gRPC)

// --- gRPC фасады (токен только в metadata, НЕ в теле) ---

type SecretBankCardListGRPCFacade struct {
	client pb.SecretBankCardServiceClient
}

func NewSecretBankCardListGRPCFacade(client pb.SecretBankCardServiceClient) *SecretBankCardListGRPCFacade {
	return &SecretBankCardListGRPCFacade{client: client}
}

func (f *SecretBankCardListGRPCFacade) List(ctx context.Context, token string) ([]models.SecretBankCardClient, error) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Теперь пустой запрос (нет token)
	resp, err := f.client.List(ctx, &pb.SecretBankCardListRequest{})
	if err != nil {
		return nil, err
	}

	var result []models.SecretBankCardClient
	for _, item := range resp.Items {
		updatedAt, err := time.Parse(time.RFC3339, item.UpdatedAt)
		if err != nil {
			return nil, errors.New("invalid updated_at format in response")
		}

		var meta *string
		if item.Meta != "" {
			meta = &item.Meta
		}

		result = append(result, models.SecretBankCardClient{
			SecretName: item.SecretName,
			Owner:      item.Owner,
			Number:     item.Number,
			Exp:        item.Exp,
			CVV:        item.Cvv,
			Meta:       meta,
			UpdatedAt:  updatedAt,
		})
	}

	return result, nil
}

type SecretBankCardGetGRPCFacade struct {
	client pb.SecretBankCardServiceClient
}

func NewSecretBankCardGetGRPCFacade(client pb.SecretBankCardServiceClient) *SecretBankCardGetGRPCFacade {
	return &SecretBankCardGetGRPCFacade{client: client}
}

func (f *SecretBankCardGetGRPCFacade) Get(ctx context.Context, token string, secretName string) (*models.SecretBankCardClient, error) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &pb.SecretBankCardGetRequest{
		SecretName: secretName,
	}

	resp, err := f.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	item := resp.Card
	updatedAt, err := time.Parse(time.RFC3339, item.UpdatedAt)
	if err != nil {
		return nil, errors.New("invalid updated_at format in response")
	}

	var meta *string
	if item.Meta != "" {
		meta = &item.Meta
	}

	return &models.SecretBankCardClient{
		SecretName: item.SecretName,
		Owner:      item.Owner,
		Number:     item.Number,
		Exp:        item.Exp,
		CVV:        item.Cvv,
		Meta:       meta,
		UpdatedAt:  updatedAt,
	}, nil
}

type SecretBankCardSaveGRPCFacade struct {
	client pb.SecretBankCardServiceClient
}

func NewSecretBankCardSaveGRPCFacade(client pb.SecretBankCardServiceClient) *SecretBankCardSaveGRPCFacade {
	return &SecretBankCardSaveGRPCFacade{client: client}
}

func (f *SecretBankCardSaveGRPCFacade) Save(ctx context.Context, token string, secret models.SecretBankCardClient) error {
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

	req := &pb.SecretBankCardSaveRequest{
		Card: &pb.SecretBankCard{
			SecretName: secret.SecretName,
			Owner:      secret.Owner,
			Number:     secret.Number,
			Exp:        secret.Exp,
			Cvv:        secret.CVV,
			Meta:       meta,
			UpdatedAt:  secret.UpdatedAt.Format(time.RFC3339),
		},
	}

	// Отправляем запрос
	_, err := f.client.Save(ctx, req)
	return err
}
