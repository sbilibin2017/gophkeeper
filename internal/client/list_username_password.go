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

// ListUsernamePasswordLocal retrieves all locally stored username-password secrets.
func ListUsernamePasswordLocal(
	ctx context.Context,
	db *sqlx.DB,
) ([]*models.UsernamePasswordAddRequest, error) {
	const query = `
		SELECT secret_name, username, password, meta
		FROM secret_username_password_request
		ORDER BY secret_name;
	`
	var results []*models.UsernamePasswordAddRequest
	err := db.SelectContext(ctx, &results, query)
	return results, err
}

// ListUsernamePasswordHTTP fetches username-password secrets via HTTP and returns full detailed responses.
func ListUsernamePasswordHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
) ([]models.UsernamePasswordResponse, error) {
	var respBody struct {
		Items []models.UsernamePasswordResponse `json:"items"`
	}

	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&respBody).
		Get("/list/username-password")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch username-password secrets: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to list username-password secrets, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return respBody.Items, nil
}

// ListUsernamePasswordGRPC fetches username-password secrets via gRPC and returns full detailed responses.
func ListUsernamePasswordGRPC(
	ctx context.Context,
	client pb.UsernamePasswordListServiceClient,
	token string,
) ([]models.UsernamePasswordResponse, error) {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	resp, err := client.List(ctxWithToken, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []models.UsernamePasswordResponse
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

		results = append(results, models.UsernamePasswordResponse{
			SecretName:  item.SecretName,
			SecretOwner: item.SecretOwner,
			Username:    item.Username,
			Password:    item.Password,
			Meta:        metaPtr,
			UpdatedAt:   updatedAt,
		})
	}

	return results, nil
}
