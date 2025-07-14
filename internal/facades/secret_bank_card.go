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

// SecretBankCardListHTTPFacade обеспечивает получение списка секретов банковских карт через HTTP.
type SecretBankCardListHTTPFacade struct {
	client *resty.Client
}

// NewBankCardListHTTPFacade создает новый SecretBankCardListHTTPFacade.
func NewBankCardListHTTPFacade(client *resty.Client) *SecretBankCardListHTTPFacade {
	return &SecretBankCardListHTTPFacade{client: client}
}

// List запрашивает список секретов банковских карт с использованием HTTP API.
// Принимает контекст и токен аутентификации.
// Возвращает слайс моделей SecretBankCardClient или ошибку.
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

// SecretBankCardListGRPCFacade обеспечивает получение списка секретов банковских карт через gRPC.
type SecretBankCardListGRPCFacade struct {
	client pb.SecretBankCardServiceClient
}

// NewBankCardListGRPCFacade создает новый SecretBankCardListGRPCFacade.
func NewBankCardListGRPCFacade(client pb.SecretBankCardServiceClient) *SecretBankCardListGRPCFacade {
	return &SecretBankCardListGRPCFacade{client: client}
}

// List запрашивает список секретов банковских карт через gRPC API.
// Принимает контекст и токен аутентификации.
// Возвращает слайс моделей SecretBankCardClient или ошибку.
func (f *SecretBankCardListGRPCFacade) List(ctx context.Context, token string) ([]models.SecretBankCardClient, error) {
	// Добавляем токен в метаданные (если используется auth через metadata)
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := f.client.ListBankCards(ctx, &pb.SecretBankCardListRequest{Token: token})
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
