package client

import (
	"context"
	"fmt"
	"net/http"

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
) ([]*models.BankCardAddRequest, error) {
	var respBody []*models.BankCardAddRequest

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

	return respBody, nil
}

// ListBankCardsGRPC uses gRPC to fetch bank cards from remote and convert them to local models.
func ListBankCardsGRPC(
	ctx context.Context,
	client pb.BankCardListServiceClient,
	token string,
) ([]*models.BankCardAddRequest, error) {

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctxWithToken, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []*models.BankCardAddRequest
	for _, pbItem := range resp.Items {
		var metaPtr *string
		if pbItem.Meta != "" {
			metaPtr = &pbItem.Meta
		}

		modelItem := &models.BankCardAddRequest{
			SecretName: pbItem.SecretName,
			Number:     pbItem.Number,
			Owner:      pbItem.Owner,
			Exp:        pbItem.Exp,
			CVV:        pbItem.Cvv,
			Meta:       metaPtr,
		}
		results = append(results, modelItem)
	}

	return results, nil
}
