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

// SecretBankCardSaveHTTP сохраняет секрет банковской карты через HTTP.
func SaveSecretBankCardHTTP(ctx context.Context, client *resty.Client, token string, secret models.SecretBankCardSaveRequest) error {
	reqBody := map[string]interface{}{
		"secret_name": secret.SecretName,
		"owner":       secret.Owner,
		"number":      secret.Number,
		"exp":         secret.Exp,
		"cvv":         secret.CVV,
	}
	if secret.Meta != nil {
		reqBody["meta"] = *secret.Meta
	}

	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetBody(reqBody).
		Post("/save/secret-bank-card")
	if err != nil {
		return fmt.Errorf("failed to send save request: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("server error: %s", resp.Status())
	}
	return nil
}

// SecretBankCardGetHTTP получает секрет банковской карты по имени через HTTP.
func GetSecretBankCardHTTP(ctx context.Context, client *resty.Client, token, secretName string) (*models.SecretBankCardGetResponse, error) {
	var secret models.SecretBankCardGetResponse
	url := fmt.Sprintf("/get/secret-bank-card/%s", secretName)
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error fetching bank card secret: %s", resp.Status())
	}
	return &secret, nil
}

// SecretBankCardListHTTP получает список всех секретов банковских карт через HTTP.
func ListSecretBankCardHTTP(ctx context.Context, client *resty.Client, token string) ([]models.SecretBankCardGetResponse, error) {
	var secrets []models.SecretBankCardGetResponse
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secrets).
		Get("/list/secret-bank-card")
	if err != nil {
		return nil, fmt.Errorf("server unavailable: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error fetching bank card secrets: %s", resp.Status())
	}
	return secrets, nil
}

// SecretBankCardSaveGRPC сохраняет секрет банковской карты через gRPC.
func SaveSecretBankCardGRPC(ctx context.Context, client pb.SecretBankCardServiceClient, token string, secret models.SecretBankCardSaveRequest) error {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	var meta string
	if secret.Meta != nil {
		meta = *secret.Meta
	} else {
		meta = ""
	}

	req := &pb.SecretBankCardSaveRequest{
		SecretName: secret.SecretName,
		Owner:      secret.Owner,
		Number:     secret.Number,
		Exp:        secret.Exp,
		Cvv:        secret.CVV,
		Meta:       meta,
	}

	_, err := client.Save(ctx, req)
	return err
}

// SecretBankCardGetGRPC получает секрет банковской карты по имени через gRPC.
func GetSecretBankCardGRPC(ctx context.Context, client pb.SecretBankCardServiceClient, token, secretName string) (*models.SecretBankCardGetResponse, error) {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &pb.SecretBankCardGetRequest{SecretName: secretName}
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
		} else {
			// Можно логировать ошибку или вернуть ошибку, если нужно
		}
	}

	return &models.SecretBankCardGetResponse{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Number:      resp.Number,
		Owner:       resp.Owner,
		Exp:         resp.Exp,
		CVV:         resp.Cvv,
		Meta:        meta,
		UpdatedAt:   updatedAt,
	}, nil
}

// SecretBankCardListGRPC получает список всех секретов банковских карт через gRPC.
func ListSecretBankCardGRPC(ctx context.Context, client pb.SecretBankCardServiceClient, token string) ([]models.SecretBankCardGetResponse, error) {
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctx, &pb.SecretBankCardListRequest{})
	if err != nil {
		return nil, err
	}

	var result []models.SecretBankCardGetResponse
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
			} else {
				// Можно логировать ошибку или игнорировать
			}
		}

		result = append(result, models.SecretBankCardGetResponse{
			SecretName:  item.SecretName,
			SecretOwner: item.SecretOwner,
			Number:      item.Number,
			Owner:       item.Owner,
			Exp:         item.Exp,
			CVV:         item.Cvv,
			Meta:        meta,
			UpdatedAt:   updatedAt,
		})
	}

	return result, nil
}

// SaveSecretBankCardRequest сохраняет или обновляет секрет с банковской картой
func SaveSecretBankCardRequest(ctx context.Context, db *sqlx.DB, card models.SecretBankCardSaveRequest) error {
	query := `
		INSERT INTO secret_bank_card_request (secret_name, number, owner, exp, cvv, meta)
		VALUES (:secret_name, :number, :owner, :exp, :cvv, :meta)
		ON CONFLICT(secret_name) DO UPDATE SET
			number = excluded.number,
			owner = excluded.owner,
			exp = excluded.exp,
			cvv = excluded.cvv,
			meta = excluded.meta;
	`

	_, err := db.NamedExecContext(ctx, query, card)
	return err
}

// GetAllSecretBankCardRequest возвращает список всех секретов банковских карт (только имена).
func GetAllSecretsBankCardRequest(ctx context.Context, db *sqlx.DB) ([]models.SecretBankCardGetRequest, error) {
	query := `
		SELECT secret_name
		FROM secret_bank_card_request;
	`

	var cards []models.SecretBankCardGetRequest
	if err := db.SelectContext(ctx, &cards, query); err != nil {
		return nil, err
	}

	return cards, nil
}

// GetSecretBankCardByNameRequest возвращает один секрет банковской карты по secret_name (без owner).
func GetSecretBankCardByNameRequest(ctx context.Context, db *sqlx.DB, secretName string) (*models.SecretBankCardGetResponse, error) {
	query := `
		SELECT secret_name, number, owner, exp, cvv, meta, updated_at
		FROM secret_bank_card_request
		WHERE secret_name = $1;
	`

	var card models.SecretBankCardGetResponse
	err := db.GetContext(ctx, &card, query, secretName)
	if err != nil {
		return nil, errors.New("secret not found or error fetching")
	}

	return &card, nil
}
