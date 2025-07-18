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
	"google.golang.org/protobuf/types/known/emptypb"
)

// ListBinaryLocal retrieves all locally stored binary secrets.
func ListBinaryLocal(
	ctx context.Context,
	db *sqlx.DB,
) ([]*models.BinaryAddRequest, error) {
	query := `
		SELECT secret_name, data, meta
		FROM secret_binary_request
		ORDER BY secret_name;
	`
	var results []*models.BinaryAddRequest
	err := db.SelectContext(ctx, &results, query)
	return results, err
}

// ListBinaryHTTP fetches binary secrets via HTTP and maps to internal models.
func ListBinaryHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
) ([]*models.BinaryResponse, error) {
	var respBody struct {
		Items []models.BinaryResponse `json:"items"`
	}

	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&respBody).
		Get("/list/binary")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch binary secrets: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list binary secrets, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	results := make([]*models.BinaryResponse, 0, len(respBody.Items))
	for i := range respBody.Items {
		results = append(results, &respBody.Items[i])
	}

	return results, nil
}

// ListBinaryGRPC fetches binary secrets via gRPC and maps them to internal models.
func ListBinaryGRPC(
	ctx context.Context,
	client pb.BinaryListServiceClient,
	token string,
) ([]*models.BinaryResponse, error) {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctxWithToken, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []*models.BinaryResponse
	for _, pbItem := range resp.Items {
		var metaPtr *string
		if pbItem.Meta != "" {
			metaPtr = &pbItem.Meta
		}

		updatedAt, err := time.Parse(time.RFC3339, pbItem.UpdatedAt)
		if err != nil {
			updatedAt = time.Time{}
		}

		result := &models.BinaryResponse{
			SecretName:  pbItem.SecretName,
			SecretOwner: pbItem.SecretOwner,
			Data:        pbItem.Data,
			Meta:        metaPtr,
			UpdatedAt:   updatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}
