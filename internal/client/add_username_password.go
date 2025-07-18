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

// AddUsernamePasswordLocal inserts or updates a UsernamePasswordAddRequest in the local DB.
func AddUsernamePasswordLocal(
	ctx context.Context,
	db *sqlx.DB,
	req models.UsernamePasswordAddRequest,
) error {
	query := `
		INSERT INTO secret_username_password_request (secret_name, username, password, meta)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (secret_name) DO UPDATE SET
			username = EXCLUDED.username,
			password = EXCLUDED.password,
			meta = EXCLUDED.meta;
	`
	_, err := db.ExecContext(ctx, query, req.SecretName, req.Username, req.Password, req.Meta)
	return err
}

// AddUsernamePasswordHTTP adds a username-password secret via HTTP with JSON body.
func AddUsernamePasswordHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	req models.UsernamePasswordAddRequest,
) error {
	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/add/username-password")

	if err != nil {
		return fmt.Errorf("failed to add username-password secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to add username-password secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return nil
}

// AddUsernamePasswordGRPC adds a username-password secret via gRPC.
func AddUsernamePasswordGRPC(
	ctx context.Context,
	client pb.UsernamePasswordAddServiceClient,
	token string,
	req models.UsernamePasswordAddRequest,
) error {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	grpcReq := &pb.UsernamePasswordAddRequest{
		SecretName: req.SecretName,
		Username:   req.Username,
		Password:   req.Password,
		Meta:       "",
	}
	if req.Meta != nil {
		grpcReq.Meta = *req.Meta
	}

	_, err := client.Add(ctxWithToken, grpcReq)
	return err
}
