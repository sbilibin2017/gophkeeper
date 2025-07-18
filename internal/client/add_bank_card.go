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
)

// AddBankCardLocal inserts a BankCardAddRequest into the local DB.
func AddBankCardLocal(
	ctx context.Context,
	db *sqlx.DB,
	req models.BankCardAddRequest,
) error {
	query := `
		INSERT INTO secret_bank_card_request (secret_name, number, owner, exp, cvv, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (secret_name) DO UPDATE SET
			number = EXCLUDED.number,
			owner = EXCLUDED.owner,
			exp = EXCLUDED.exp,
			cvv = EXCLUDED.cvv,
			meta = EXCLUDED.meta;
	`
	_, err := db.ExecContext(ctx, query, req.SecretName, req.Number, req.Owner, req.Exp, req.CVV, req.Meta)
	return err
}

// AddBankCardHTTP adds a bank card secret via HTTP with JSON body.
func AddBankCardHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	req models.BankCardAddRequest,
) error {
	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/add/bank-card")

	if err != nil {
		return fmt.Errorf("failed to add bank card secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to add bank card secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return nil
}

// AddBankCardGRPC adds a bank card secret via gRPC.
func AddBankCardGRPC(
	ctx context.Context,
	client pb.BankCardAddServiceClient,
	token string,
	req models.BankCardAddRequest,
) error {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	grpcReq := &pb.BankCardAddRequest{
		SecretName: req.SecretName,
		Number:     req.Number,
		Owner:      req.Owner,
		Exp:        req.Exp,
		Cvv:        req.CVV,
		Meta:       "",
	}
	if req.Meta != nil {
		grpcReq.Meta = *req.Meta
	}

	_, err := client.Add(ctxWithToken, grpcReq)
	return err
}
