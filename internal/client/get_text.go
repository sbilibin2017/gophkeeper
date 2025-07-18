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
)

// GetTextLocal retrieves a single text secret by secretName from local DB.
func GetTextLocal(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) (*models.TextAddRequest, error) {
	const query = `
		SELECT secret_name, content, meta
		FROM secret_text_request
		WHERE secret_name = ?;
	`

	var result models.TextAddRequest
	err := db.GetContext(ctx, &result, query, secretName)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTextHTTP fetches a single text secret by secretName via HTTP using path parameter.
func GetTextHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretName string,
) (*models.TextResponse, error) {
	var respBody struct {
		Item models.TextResponse `json:"item"`
	}

	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&respBody).
		Get(fmt.Sprintf("/get/text/%s", secretName)) // path param here
	if err != nil {
		return nil, fmt.Errorf("failed to fetch text secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get text secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return &respBody.Item, nil
}

// GetTextGRPC fetches a single text secret by secretName via gRPC.
func GetTextGRPC(
	ctx context.Context,
	client pb.TextGetServiceClient, // note: this is the client for Get service
	token string,
	secretName string,
) (*models.TextResponse, error) {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	req := &pb.TextGetRequest{
		SecretName: secretName,
	}
	resp, err := client.Get(ctxWithToken, req)
	if err != nil {
		return nil, err
	}

	var metaPtr *string
	if resp.Meta != "" {
		metaPtr = &resp.Meta
	}

	var updatedAt time.Time
	if resp.UpdatedAt != "" {
		updatedAt, err = time.Parse(time.RFC3339, resp.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse UpdatedAt: %w", err)
		}
	}

	return &models.TextResponse{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Content:     resp.Content,
		Meta:        metaPtr,
		UpdatedAt:   updatedAt,
	}, nil
}
