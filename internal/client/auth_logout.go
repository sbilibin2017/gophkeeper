package client

import (
	"context"
	"fmt"
	"net/http"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// LogoutHTTP sends an HTTP logout request.
//
// It takes a context, a Resty HTTP client, and a LogoutRequest model.
// Returns an error if the logout failed.
func LogoutHTTP(ctx context.Context, client *resty.Client, req *models.LogoutRequest) error {
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", req.Token).
		Post("/logout")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("logout failed: %s", resp.Status())
	}

	return nil
}

// LogoutGRPC performs user logout using a gRPC client.
//
// It takes a context, a gRPC LogoutServiceClient, and a LogoutRequest model.
// Returns an error if the logout failed.
func LogoutGRPC(
	ctx context.Context,
	client pb.LogoutServiceClient,
	req *models.LogoutRequest,
) error {
	_, err := client.Logout(ctx, &pb.LogoutRequest{
		Token: req.Token,
	})
	if err != nil {
		return fmt.Errorf("grpc logout call failed: %w", err)
	}
	return nil
}
