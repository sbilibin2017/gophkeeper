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

// AddBinaryLocal inserts a BinaryAddRequest into the local DB.
func AddBinaryLocal(
	ctx context.Context,
	db *sqlx.DB,
	req models.BinaryAddRequest,
) error {
	query := `
		INSERT INTO secret_binary_request (secret_name, data, meta)
		VALUES ($1, $2, $3)
		ON CONFLICT (secret_name) DO UPDATE SET
			data = EXCLUDED.data,
			meta = EXCLUDED.meta;
	`
	_, err := db.ExecContext(ctx, query, req.SecretName, req.Data, req.Meta)
	return err
}

// AddBinaryHTTP adds a binary secret via HTTP with JSON body.
func AddBinaryHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	req models.BinaryAddRequest,
) error {
	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/add/binary")

	if err != nil {
		return fmt.Errorf("failed to add binary secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to add binary secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return nil
}

// AddBinaryGRPC adds a binary secret via gRPC.
func AddBinaryGRPC(
	ctx context.Context,
	client pb.BinaryAddServiceClient,
	token string,
	req models.BinaryAddRequest,
) error {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	grpcReq := &pb.BinaryAddRequest{
		SecretName: req.SecretName,
		Data:       req.Data,
		Meta:       "",
	}
	if req.Meta != nil {
		grpcReq.Meta = *req.Meta
	}

	_, err := client.Add(ctxWithToken, grpcReq)
	return err
}
