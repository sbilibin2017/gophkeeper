package auth

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
)

// LoginHTTPFacade handles HTTP-based user login.
type LoginHTTPFacade struct {
	client *resty.Client
}

// NewLoginHTTPFacade creates a new instance of LoginHTTPFacade.
func NewLoginHTTPFacade(client *resty.Client) *LoginHTTPFacade {
	return &LoginHTTPFacade{client: client}
}

// Login performs login via HTTP POST to "/login".
func (f *LoginHTTPFacade) Login(ctx context.Context, req *models.AuthRequest) (*models.AuthResponse, error) {
	var authResp models.AuthResponse

	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&authResp).
		Post("/login")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("login failed with status: %s", resp.Status())
	}

	return &authResp, nil
}

// LoginGRPCFacade handles gRPC-based user login.
type LoginGRPCFacade struct {
	client pb.AuthServiceClient
}

// NewLoginGRPCFacade creates a new instance of LoginGRPCFacade.
func NewLoginGRPCFacade(client pb.AuthServiceClient) *LoginGRPCFacade {
	return &LoginGRPCFacade{client: client}
}

// Login performs login via gRPC call.
func (f *LoginGRPCFacade) Login(ctx context.Context, req *models.AuthRequest) (*models.AuthResponse, error) {
	grpcReq := &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}

	resp, err := f.client.Login(ctx, grpcReq)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{Token: resp.Token}, nil
}
