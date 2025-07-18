package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc/metadata"
)

// GetBankCardLocal retrieves a single bank card secret by secretName from local DB.
func GetBankCardLocal(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) (*models.BankCardAddRequest, error) {
	const query = `
		SELECT secret_name, number, owner, exp, cvv, meta
		FROM secret_bank_card_request
		WHERE secret_name = ?;
	`

	var result models.BankCardAddRequest
	err := db.GetContext(ctx, &result, query, secretName)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBankCardHTTP fetches a single bank card secret by secretName via HTTP using path parameter.
func GetBankCardHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretName string,
) (*models.BankCardResponse, error) {
	var respBody struct {
		Item models.BankCardResponse `json:"item"`
	}

	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&respBody).
		Get(fmt.Sprintf("/get/bank-card/%s", secretName)) // path param
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bank card secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get bank card secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return &respBody.Item, nil
}

// GetBankCardGRPC fetches a single bank card secret by secretName via gRPC.
func GetBankCardGRPC(
	ctx context.Context,
	client pb.BankCardGetServiceClient, // gRPC client for BankCardGetService
	token string,
	secretName string,
) (*models.BankCardResponse, error) {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	req := &pb.BankCardGetRequest{
		SecretName: secretName,
	}
	resp, err := client.Get(ctxWithToken, req)
	if err != nil {
		return nil, err
	}

	var metaPtr *string
	if resp.Meta != "" {
		metaPtr = &resp.Meta
	}

	var updatedAt time.Time
	if resp.UpdatedAt != "" {
		updatedAt, err = time.Parse(time.RFC3339, resp.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse UpdatedAt: %w", err)
		}
	}

	return &models.BankCardResponse{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Number:      resp.Number,
		Owner:       resp.Owner,
		Exp:         resp.Exp,
		CVV:         resp.Cvv,
		Meta:        metaPtr,
		UpdatedAt:   updatedAt,
	}, nil
}
