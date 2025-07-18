package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// RegisterHTTP sends an HTTP registration request to the server.
func RegisterHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.RegisterRequest,
) error {
	httpResp, err := client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/register")
	if err != nil {
		return fmt.Errorf("failed to send HTTP register request: %w", err)
	}

	if httpResp.StatusCode() != http.StatusOK {
		return fmt.Errorf("registration failed with status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return nil
}

// RegisterGRPC performs user registration using a gRPC client.
func RegisterGRPC(
	ctx context.Context,
	client pb.RegisterServiceClient,
	req *models.RegisterRequest,
) error {
	pbReq := &pb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	}

	_, err := client.Register(ctx, pbReq)
	if err != nil {
		return fmt.Errorf("grpc register call failed: %w", err)
	}

	return nil
}
