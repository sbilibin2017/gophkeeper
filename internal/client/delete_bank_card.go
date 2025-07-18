package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc/metadata"
)

// DeleteBankCardLocal deletes a bank card secret by secretName from local DB.
func DeleteBankCardLocal(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) error {
	const query = `
		DELETE FROM secret_bank_card_request
		WHERE secret_name = ?;
	`

	_, err := db.ExecContext(ctx, query, secretName)
	return err
}

// DeleteBankCardHTTP deletes a bank card secret by secretName via HTTP using path parameter.
func DeleteBankCardHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretName string,
) error {
	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		Delete(fmt.Sprintf("/delete/bank-card/%s", secretName)) // path param

	if err != nil {
		return fmt.Errorf("failed to delete bank card secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete bank card secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return nil
}

// DeleteBankCardGRPC deletes a bank card secret by secretName via gRPC.
func DeleteBankCardGRPC(
	ctx context.Context,
	client pb.BankCardDeleteServiceClient,
	token string,
	secretName string,
) error {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	req := &pb.BankCardDeleteRequest{
		SecretName: secretName,
	}

	_, err := client.Delete(ctxWithToken, req)
	return err
}
