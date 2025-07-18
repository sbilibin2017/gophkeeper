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

// GetBinaryLocal retrieves a single binary secret by secretName from local DB.
func GetBinaryLocal(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) (*models.BinaryAddRequest, error) {
	const query = `
		SELECT secret_name, data, meta
		FROM secret_binary_request
		WHERE secret_name = ?;
	`

	var result models.BinaryAddRequest
	err := db.GetContext(ctx, &result, query, secretName)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBinaryHTTP fetches a single binary secret by secretName via HTTP using path parameter.
func GetBinaryHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretName string,
) (*models.BinaryResponse, error) {
	var respBody struct {
		Item models.BinaryResponse `json:"item"`
	}

	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&respBody).
		Get(fmt.Sprintf("/get/binary/%s", secretName)) // path param here
	if err != nil {
		return nil, fmt.Errorf("failed to fetch binary secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get binary secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return &respBody.Item, nil
}

// GetBinaryGRPC fetches a single binary secret by secretName via gRPC.
func GetBinaryGRPC(
	ctx context.Context,
	client pb.BinaryGetServiceClient, // gRPC client for BinaryGetService
	token string,
	secretName string,
) (*models.BinaryResponse, error) {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	req := &pb.BinaryGetRequest{
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

	return &models.BinaryResponse{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Data:        resp.Data,
		Meta:        metaPtr,
		UpdatedAt:   updatedAt,
	}, nil
}
