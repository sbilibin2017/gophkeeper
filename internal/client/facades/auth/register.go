package auth

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/client/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
)

// AuthRegisterHTTPFacade handles HTTP-based registration.
type RegisterHTTPFacade struct {
	client *resty.Client
}

// NewAuthRegisterHTTPFacade returns a new instance of AuthRegisterHTTPFacade.
func NewRegisterHTTPFacade(client *resty.Client) *RegisterHTTPFacade {
	return &RegisterHTTPFacade{client: client}
}

// Register performs registration via HTTP.
func (f *RegisterHTTPFacade) Register(
	ctx context.Context,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	var authResp models.AuthResponse

	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&authResp).
		Post("/register")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("registration failed with status: %s", resp.Status())
	}

	return &authResp, nil
}

// AuthRegisterGRPCFacade handles gRPC-based registration.
type RegisterGRPCFacade struct {
	client pb.AuthServiceClient
}

// NewAuthRegisterGRPCFacade returns a new instance of AuthRegisterGRPCFacade.
func NewRegisterGRPCFacade(client pb.AuthServiceClient) *RegisterGRPCFacade {
	return &RegisterGRPCFacade{client: client}
}

// Register performs registration via gRPC.
func (f *RegisterGRPCFacade) Register(
	ctx context.Context,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	grpcReq := &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}

	resp, err := f.client.Register(ctx, grpcReq)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{Token: resp.Token}, nil
}
