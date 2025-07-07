package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// HTTPLoginService implements authentication via HTTP API.
type HTTPLoginService struct {
	client *resty.Client
}

// NewHTTPLoginService creates a new HTTPLoginService with the given HTTP client.
func NewHTTPLoginService(client *resty.Client) *HTTPLoginService {
	return &HTTPLoginService{client: client}
}

// Login sends an HTTP POST request to authenticate the user.
// Returns an error if the request fails or the server returns a status other than 200 OK.
func (l *HTTPLoginService) Login(ctx context.Context, creds *models.Credentials) error {
	resp, err := l.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(creds).
		Post("/login")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("login failed: %s", resp.String())
	}

	return nil
}

// GRPCLoginService implements authentication via gRPC service.
type GRPCLoginService struct {
	client pb.LoginServiceClient
}

// NewGRPCLoginService creates a new GRPCLoginService with the given gRPC client.
func NewGRPCLoginService(client pb.LoginServiceClient) *GRPCLoginService {
	return &GRPCLoginService{client: client}
}

// Login sends a gRPC request to authenticate the user.
// Returns an error if the request fails.
func (l *GRPCLoginService) Login(ctx context.Context, creds *models.Credentials) error {
	_, err := l.client.Login(ctx, &pb.LoginRequest{
		Username: creds.Username,
		Password: creds.Password,
	})
	if err != nil {
		return err
	}

	return nil
}
