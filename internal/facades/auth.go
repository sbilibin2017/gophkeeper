package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
)

// AuthHTTPFacade provides HTTP-based authentication methods.
type AuthHTTPFacade struct {
	client *resty.Client
}

// NewAuthHTTPFacade creates a new AuthHTTPFacade with the given Resty client.
func NewAuthHTTPFacade(client *resty.Client) *AuthHTTPFacade {
	return &AuthHTTPFacade{client: client}
}

func (a *AuthHTTPFacade) Register(
	ctx context.Context,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	resp, err := a.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/register")
	if err != nil {
		return nil, fmt.Errorf("register request failed: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("register request returned error: %s", resp.Status())
	}

	authHeader := resp.Header().Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header missing in response")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return nil, fmt.Errorf("authorization header format invalid")
	}

	token := authHeader[len(bearerPrefix):]

	return &models.AuthResponse{Token: token}, nil
}

func (a *AuthHTTPFacade) Login(
	ctx context.Context,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	resp, err := a.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/login")
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("login request returned error: %s", resp.Status())
	}

	authHeader := resp.Header().Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header missing in response")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return nil, fmt.Errorf("authorization header format invalid")
	}

	token := authHeader[len(bearerPrefix):]

	return &models.AuthResponse{Token: token}, nil
}

// AuthGRPCFacade provides gRPC-based authentication methods.
type AuthGRPCFacade struct {
	client pb.AuthServiceClient
}

// NewAuthGRPCFacade creates a new AuthGRPCFacade with the given gRPC client connection.
func NewAuthGRPCFacade(conn *grpc.ClientConn) *AuthGRPCFacade {
	return &AuthGRPCFacade{
		client: pb.NewAuthServiceClient(conn),
	}
}

// Register sends a registration request over gRPC and returns an AuthResponse or an error.
func (a *AuthGRPCFacade) Register(
	ctx context.Context,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	resp, err := a.client.Register(ctx, &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &models.AuthResponse{Token: resp.Token}, nil
}

// Login sends a login request over gRPC and returns an AuthResponse or an error.
func (a *AuthGRPCFacade) Login(
	ctx context.Context,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	resp, err := a.client.Login(ctx, &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &models.AuthResponse{Token: resp.Token}, nil
}
