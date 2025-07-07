package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// HTTPRegisterService implements registration via HTTP API.
type HTTPRegisterService struct {
	client *resty.Client
}

// NewHTTPRegisterService creates a new HTTPRegisterService with the specified HTTP client.
func NewHTTPRegisterService(client *resty.Client) *HTTPRegisterService {
	return &HTTPRegisterService{client: client}
}

// Register sends an HTTP POST request to register the user.
// Returns an error if the request failed or the server returned a status other than 200 OK.
func (r *HTTPRegisterService) Register(
	ctx context.Context,
	creds *models.Credentials,
) error {
	resp, err := r.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(creds).
		Post("/register")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("registration failed: %s", resp.String())
	}

	return nil
}

// GRPCRegisterService implements registration via gRPC service.
type GRPCRegisterService struct {
	client pb.RegisterServiceClient
}

// NewGRPCRegisterService creates a new GRPCRegisterService with the specified gRPC client.
func NewGRPCRegisterService(client pb.RegisterServiceClient) *GRPCRegisterService {
	return &GRPCRegisterService{client: client}
}

// Register sends a gRPC request to register the user.
// Returns an error if the request failed.
func (r *GRPCRegisterService) Register(ctx context.Context, creds *models.Credentials) error {
	_, err := r.client.Register(ctx, &pb.RegisterRequest{
		Username: creds.Username,
		Password: creds.Password,
	})
	if err != nil {
		return err
	}

	return nil
}
