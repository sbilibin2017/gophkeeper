package client

import (
	"context"
	"fmt"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// AuthHTTP sends an HTTP authentication request using models.AuthRequest and models.AuthResponse.
// It posts the request to the "/auth" endpoint (update this path to match your actual API).
//
// Returns the AuthResponse containing the authentication token if successful,
// or an error if the request failed or the status code is not 200.
func AuthHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	resp := &models.AuthResponse{}

	httpResp, err := client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(resp).
		Post("/auth")
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP auth request: %w", err)
	}

	if httpResp.StatusCode() != 200 {
		return nil, fmt.Errorf("auth failed with status %d: %s", httpResp.StatusCode(), httpResp.String())
	}

	return resp, nil
}

// AuthGRPC calls the Auth method of the gRPC AuthService using protobuf request and response.
// It converts the models.AuthRequest to protobuf.AuthRequest,
// calls the service, and then converts the protobuf.AuthResponse back to models.AuthResponse.
//
// Returns the AuthResponse containing the authentication token if successful,
// or an error if the gRPC call failed.
func AuthGRPC(
	ctx context.Context,
	client pb.AuthServiceClient,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	pbReq := &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}

	pbResp, err := client.Auth(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("grpc auth call failed: %w", err)
	}

	resp := &models.AuthResponse{
		Token: pbResp.Token,
	}

	return resp, nil
}
