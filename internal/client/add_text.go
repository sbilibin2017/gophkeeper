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

// AddTextLocal inserts a TextAddRequest into the local DB.
func AddTextLocal(
	ctx context.Context,
	db *sqlx.DB,
	req models.TextAddRequest,
) error {
	query := `
		INSERT INTO secret_text_request (secret_name, content, meta)
		VALUES ($1, $2, $3)
		ON CONFLICT (secret_name) DO UPDATE SET
			content = EXCLUDED.content,
			meta = EXCLUDED.meta;
	`
	_, err := db.ExecContext(ctx, query, req.SecretName, req.Content, req.Meta)
	return err
}

// AddTextHTTP adds a text secret via HTTP with JSON body.
func AddTextHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	req models.TextAddRequest,
) error {
	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/add/text")

	if err != nil {
		return fmt.Errorf("failed to add text secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to add text secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return nil
}

// AddTextGRPC adds a text secret via gRPC.
func AddTextGRPC(
	ctx context.Context,
	client pb.TextAddServiceClient,
	token string,
	req models.TextAddRequest,
) error {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	grpcReq := &pb.TextAddRequest{
		SecretName: req.SecretName,
		Content:    req.Content,
		Meta:       "",
	}
	if req.Meta != nil {
		grpcReq.Meta = *req.Meta
	}

	_, err := client.Add(ctxWithToken, grpcReq)
	return err
}
