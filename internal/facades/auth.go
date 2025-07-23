package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AuthHTTPFacade provides HTTP client methods for authentication.
type AuthHTTPFacade struct {
	client *resty.Client
}

// Register sends user registration data to the HTTP API and returns AuthResponse.
func (f *AuthHTTPFacade) Register(
	ctx context.Context, req *models.AuthRequest,
) (*models.AuthResponse, error) {
	result := &models.AuthResponse{}
	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(result).
		Post("/auth/register")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to register: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return result, nil
}

// Login sends user login data to the HTTP API and returns AuthResponse.
func (f *AuthHTTPFacade) Login(
	ctx context.Context, req *models.AuthRequest,
) (*models.AuthResponse, error) {
	result := &models.AuthResponse{}
	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(result).
		Post("/auth/login")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to login: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return result, nil
}

// Logout calls the logout HTTP endpoint.
func (f *AuthHTTPFacade) Logout(ctx context.Context) error {
	resp, err := f.client.R().
		SetContext(ctx).
		Post("/auth/logout")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to logout: status %d, body: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// AuthGRPCFacade provides gRPC client methods for authentication.
type AuthGRPCFacade struct {
	client pb.AuthServiceClient
}

// NewAuthGRPCFacade creates a new gRPC facade for AuthService.
func NewAuthGRPCFacade(conn *grpc.ClientConn) *AuthGRPCFacade {
	return &AuthGRPCFacade{
		client: pb.NewAuthServiceClient(conn),
	}
}

// Register calls the Register RPC and returns AuthResponse.
func (f *AuthGRPCFacade) Register(
	ctx context.Context, req *models.AuthRequest,
) (*models.AuthResponse, error) {
	rpcReq := &pb.AuthRequest{
		Username: req.Login,
		Password: req.Password,
	}

	resp, err := f.client.Register(ctx, rpcReq)
	if err != nil {
		return nil, fmt.Errorf("grpc Register failed: %w", err)
	}

	return &models.AuthResponse{Token: resp.Token}, nil
}

// Login calls the Login RPC and returns AuthResponse.
func (f *AuthGRPCFacade) Login(
	ctx context.Context, req *models.AuthRequest,
) (*models.AuthResponse, error) {
	rpcReq := &pb.AuthRequest{
		Username: req.Login,
		Password: req.Password,
	}

	resp, err := f.client.Login(ctx, rpcReq)
	if err != nil {
		return nil, fmt.Errorf("grpc Login failed: %w", err)
	}

	return &models.AuthResponse{Token: resp.Token}, nil
}

// Logout calls the Logout RPC.
func (f *AuthGRPCFacade) Logout(ctx context.Context) error {
	_, err := f.client.Logout(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("grpc Logout failed: %w", err)
	}
	return nil
}
