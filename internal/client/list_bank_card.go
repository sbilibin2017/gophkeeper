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
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListBankCardsLocal retrieves all locally stored bank card secrets.
func ListBankCardsLocal(
	ctx context.Context,
	db *sqlx.DB,
) ([]*models.BankCardAddRequest, error) {
	query := `
		SELECT secret_name, number, owner, exp, cvv, meta
		FROM secret_bank_card_request
		ORDER BY secret_name;
	`
	var results []*models.BankCardAddRequest
	err := db.SelectContext(ctx, &results, query)
	return results, err
}

// ListBankCardsHTTP fetches bank cards via HTTP REST API and converts them to internal models.
func ListBankCardsHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
) ([]*models.BankCardResponse, error) {
	var respBody struct {
		Items []models.BankCardResponse `json:"items"`
	}

	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&respBody).
		Get("/list/bank-card")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bank cards: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list bank cards, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	// Convert to slice of pointers
	results := make([]*models.BankCardResponse, 0, len(respBody.Items))
	for i := range respBody.Items {
		results = append(results, &respBody.Items[i])
	}

	return results, nil
}

// ListBankCardsGRPC uses gRPC to fetch bank cards from remote and convert them to local models.
func ListBankCardsGRPC(
	ctx context.Context,
	client pb.BankCardListServiceClient,
	token string,
) ([]*models.BankCardResponse, error) {

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctxWithToken, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []*models.BankCardResponse
	for _, pbItem := range resp.Items {
		var metaPtr *string
		if pbItem.Meta != "" {
			metaPtr = &pbItem.Meta
		}

		updatedAt, err := time.Parse(time.RFC3339, pbItem.UpdatedAt)
		if err != nil {
			updatedAt = time.Time{}
		}

		result := &models.BankCardResponse{
			SecretName:  pbItem.SecretName,
			SecretOwner: pbItem.SecretOwner,
			Number:      pbItem.Number,
			Owner:       pbItem.Owner,
			Exp:         pbItem.Exp,
			CVV:         pbItem.Cvv,
			Meta:        metaPtr,
			UpdatedAt:   updatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}
