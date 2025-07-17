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

// ListTextLocal retrieves all locally stored text secrets.
func ListTextLocal(
	ctx context.Context,
	db *sqlx.DB,
) ([]*models.TextAddRequest, error) {
	query := `
		SELECT secret_name, content, meta
		FROM secret_text_request
		ORDER BY secret_name;
	`
	var results []*models.TextAddRequest
	err := db.SelectContext(ctx, &results, query)
	return results, err
}

// ListTextHTTP fetches text secrets via HTTP and maps them to internal models.
func ListTextHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
) ([]*models.TextAddRequest, error) {
	var respBody struct {
		Items []models.TextResponse `json:"items"`
	}

	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&respBody).
		Get("/list/text")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch text secrets: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list text secrets, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	var results []*models.TextAddRequest
	for _, item := range respBody.Items {
		results = append(results, &models.TextAddRequest{
			SecretName: item.SecretName,
			Content:    item.Content,
			Meta:       item.Meta,
		})
	}

	return results, nil
}

// ListTextGRPC fetches text secrets via gRPC and maps them to internal models.
func ListTextGRPC(
	ctx context.Context,
	client pb.TextListServiceClient,
	token string,
) ([]*models.TextAddRequest, error) {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctxWithToken, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []*models.TextAddRequest
	for _, item := range resp.Items {
		var metaPtr *string
		if item.Meta != "" {
			metaPtr = &item.Meta
		}
		results = append(results, &models.TextAddRequest{
			SecretName: item.SecretName,
			Content:    item.Content,
			Meta:       metaPtr,
		})
	}

	return results, nil
}
