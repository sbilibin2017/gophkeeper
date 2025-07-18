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

// DeleteTextLocal deletes a text secret by secretName from local DB.
func DeleteTextLocal(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) error {
	const query = `
		DELETE FROM secret_text_request
		WHERE secret_name = ?;
	`

	_, err := db.ExecContext(ctx, query, secretName)
	return err
}

// DeleteTextHTTP deletes a text secret by secretName via HTTP using path parameter.
func DeleteTextHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretName string,
) error {
	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		Delete(fmt.Sprintf("/delete/text/%s", secretName)) // path param

	if err != nil {
		return fmt.Errorf("failed to delete text secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete text secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return nil
}

// DeleteTextGRPC deletes a text secret by secretName via gRPC.
func DeleteTextGRPC(
	ctx context.Context,
	client pb.TextDeleteServiceClient,
	token string,
	secretName string,
) error {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	req := &pb.TextDeleteRequest{
		SecretName: secretName,
	}

	_, err := client.Delete(ctxWithToken, req)
	return err
}
