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
) ([]*models.BinaryAddRequest, error) {
	var respBody models.BinaryListResponse

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

	var results []*models.BinaryAddRequest
	for _, item := range respBody.Items {
		result := &models.BinaryAddRequest{
			SecretName: item.SecretName,
			Data:       item.Data,
			Meta:       item.Meta,
		}
		results = append(results, result)
	}

	return results, nil
}

// ListBinaryGRPC fetches binary secrets via gRPC and maps them to internal models.
func ListBinaryGRPC(
	ctx context.Context,
	client pb.BinaryListServiceClient,
	token string,
) ([]*models.BinaryAddRequest, error) {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctxWithToken, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []*models.BinaryAddRequest
	for _, pbItem := range resp.Items {
		var metaPtr *string
		if pbItem.Meta != "" {
			metaPtr = &pbItem.Meta
		}

		result := &models.BinaryAddRequest{
			SecretName: pbItem.SecretName,
			Data:       pbItem.Data,
			Meta:       metaPtr,
		}
		results = append(results, result)
	}

	return results, nil
}
