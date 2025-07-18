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

// DeleteUsernamePasswordLocal deletes a username-password secret by secretName from local DB.
func DeleteUsernamePasswordLocal(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) error {
	const query = `
		DELETE FROM secret_username_password_request
		WHERE secret_name = ?;
	`

	_, err := db.ExecContext(ctx, query, secretName)
	return err
}

// DeleteUsernamePasswordHTTP deletes a username-password secret by secretName via HTTP using path parameter.
func DeleteUsernamePasswordHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretName string,
) error {
	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		Delete(fmt.Sprintf("/delete/username-password/%s", secretName)) // path param

	if err != nil {
		return fmt.Errorf("failed to delete username-password secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to delete username-password secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return nil
}

// DeleteUsernamePasswordGRPC deletes a username-password secret by secretName via gRPC.
func DeleteUsernamePasswordGRPC(
	ctx context.Context,
	client pb.UsernamePasswordDeleteServiceClient,
	token string,
	secretName string,
) error {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	req := &pb.UsernamePasswordDeleteRequest{
		SecretName: secretName,
	}

	_, err := client.Delete(ctxWithToken, req)
	return err
}
