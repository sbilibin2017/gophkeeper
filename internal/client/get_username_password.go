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

// GetUsernamePasswordLocal retrieves a single username-password secret by secretName from local DB.
func GetUsernamePasswordLocal(
	ctx context.Context,
	db *sqlx.DB,
	secretName string,
) (*models.UsernamePasswordAddRequest, error) {
	const query = `
		SELECT secret_name, username, password, meta
		FROM secret_username_password_request
		WHERE secret_name = ?;
	`

	var result models.UsernamePasswordAddRequest
	err := db.GetContext(ctx, &result, query, secretName)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUsernamePasswordHTTP fetches a single username-password secret by secretName via HTTP using path parameter.
func GetUsernamePasswordHTTP(
	ctx context.Context,
	client *resty.Client,
	token string,
	secretName string,
) (*models.UsernamePasswordResponse, error) {
	var respBody struct {
		Item models.UsernamePasswordResponse `json:"item"`
	}

	httpResp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&respBody).
		Get(fmt.Sprintf("/get/username-password/%s", secretName)) // path param here
	if err != nil {
		return nil, fmt.Errorf("failed to fetch username-password secret: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get username-password secret, status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return &respBody.Item, nil
}

// GetUsernamePasswordGRPC fetches a single username-password secret by secretName via gRPC.
func GetUsernamePasswordGRPC(
	ctx context.Context,
	client pb.UsernamePasswordGetServiceClient, // <--- Use GetServiceClient here
	token string,
	secretName string,
) (*models.UsernamePasswordResponse, error) {
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctxWithToken := metadata.NewOutgoingContext(ctx, md)

	req := &pb.UsernamePasswordGetRequest{
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

	return &models.UsernamePasswordResponse{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Username:    resp.Username,
		Password:    resp.Password,
		Meta:        metaPtr,
		UpdatedAt:   updatedAt,
	}, nil
}
