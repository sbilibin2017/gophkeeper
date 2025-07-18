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

// ListTextHTTP fetches text secrets via HTTP and returns full detailed responses.
func ListTextHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
) ([]models.TextResponse, error) {
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

	return respBody.Items, nil
}

// ListTextGRPC fetches text secrets via gRPC and returns full detailed responses.
func ListTextGRPC(
	ctx context.Context,
	client pb.TextListServiceClient,
	token string,
) ([]models.TextResponse, error) {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctxWithToken, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []models.TextResponse
	for _, item := range resp.Items {
		var metaPtr *string
		if item.Meta != "" {
			metaPtr = &item.Meta
		}

		// Parse UpdatedAt string to time.Time
		var updatedAt time.Time
		if item.UpdatedAt != "" {
			updatedAt, err = time.Parse(time.RFC3339, item.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to parse UpdatedAt: %w", err)
			}
		}

		results = append(results, models.TextResponse{
			SecretName:  item.SecretName,
			SecretOwner: item.SecretOwner,
			Content:     item.Content,
			Meta:        metaPtr,
			UpdatedAt:   updatedAt,
		})
	}

	return results, nil
}
