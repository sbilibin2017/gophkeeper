package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
)

// AuthHTTPFacade wraps an HTTP client to communicate with authentication endpoints.
type AuthHTTPFacade struct {
	client *resty.Client
}

// NewAuthHTTPFacade creates a new AuthHTTPFacade with the given Resty client.
func NewAuthHTTPFacade(client *resty.Client) (*AuthHTTPFacade, error) {
	return &AuthHTTPFacade{client: client}, nil
}

// Register sends a registration request to the server and returns the response.
func (f *AuthHTTPFacade) Register(
	ctx context.Context,
	req *models.UserRegisterRequest,
) (*models.UserRegisterResponse, error) {

	var resp models.UserRegisterResponse
	r, err := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/register/")

	if err != nil {
		return nil, fmt.Errorf("register request failed: %w", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("register request returned status %d: %s", r.StatusCode(), r.String())
	}

	return &resp, nil
}

// Login sends a login request to the server and returns the response.
func (f *AuthHTTPFacade) Login(
	ctx context.Context,
	req *models.UserLoginRequest,
) (*models.UserLoginResponse, error) {

	var resp models.UserLoginResponse
	r, err := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&resp).
		Post("/login/")

	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}

	if r.IsError() {
		return nil, fmt.Errorf("login request returned status %d: %s", r.StatusCode(), r.String())
	}

	return &resp, nil
}

// AuthGRPCFacade is a wrapper for the gRPC AuthService client.
type AuthGRPCFacade struct {
	client pb.AuthServiceClient
}

// NewAuthGRPCFacade creates a new AuthGRPCFacade with the given gRPC client connection.
func NewAuthGRPCFacade(conn *grpc.ClientConn) (*AuthGRPCFacade, error) {
	client := pb.NewAuthServiceClient(conn)
	return &AuthGRPCFacade{client: client}, nil
}

// Register calls the Register RPC method on the AuthService,
// converting between internal models and protobuf messages.
func (f *AuthGRPCFacade) Register(
	ctx context.Context,
	req *models.UserRegisterRequest,
) (*models.UserRegisterResponse, error) {

	pbReq := &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}

	pbResp, err := f.client.Register(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("grpc register failed: %w", err)
	}

	resp := &models.UserRegisterResponse{
		Token: pbResp.Token,
	}

	return resp, nil
}

// Login calls the Login RPC method on the AuthService,
// converting between internal models and protobuf messages.
func (f *AuthGRPCFacade) Login(
	ctx context.Context,
	req *models.UserLoginRequest,
) (*models.UserLoginResponse, error) {

	pbReq := &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}

	pbResp, err := f.client.Login(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("grpc login failed: %w", err)
	}

	resp := &models.UserLoginResponse{
		Token: pbResp.Token,
	}

	return resp, nil
}
