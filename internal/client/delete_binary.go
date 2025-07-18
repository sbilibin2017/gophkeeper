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

// DeleteBinaryLocal deletes a binary secret by secretName from local DB.
func DeleteBinaryLocal(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) error {
	const query = `
		DELETE FROM secret_binary_request
		WHERE secret_name = ?;
	`

	_, err := db.ExecContext(ctx, query, secretName)
	return err
}

// DeleteBinaryHTTP deletes a binary secret by secretName via HTTP using path parameter.
func DeleteBinaryHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretName string,
) error {
	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		Delete(fmt.Sprintf("/delete/binary/%s", secretName)) // path param

	if err != nil {
		return fmt.Errorf("failed to delete binary secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete binary secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return nil
}

// DeleteBinaryGRPC deletes a binary secret by secretName via gRPC.
func DeleteBinaryGRPC(
	ctx context.Context,
	client pb.BinaryDeleteServiceClient,
	token string,
	secretName string,
) error {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	req := &pb.BinaryDeleteRequest{
		SecretName: secretName,
	}

	_, err := client.Delete(ctxWithToken, req)
	return err
}
