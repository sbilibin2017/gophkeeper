package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/inernal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
)

// AuthHTTP provides HTTP methods for auth operations.
type AuthHTTP struct {
	client *resty.Client
}

// NewAuthHTTP constructs AuthHTTP with given Resty client.
func NewAuthHTTP(client *resty.Client) *AuthHTTP {
	return &AuthHTTP{client: client}
}

// Register sends a registration request to the server over HTTP.
func (h *AuthHTTP) Register(ctx context.Context, req *models.AuthRegisterRequest) (*models.AuthResponse, error) {
	var resp models.AuthResponse
	r, err := h.client.R().
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

// Login sends a login request to the server over HTTP.
func (h *AuthHTTP) Login(ctx context.Context, req *models.AuthLoginRequest) (*models.AuthResponse, error) {
	var resp models.AuthResponse
	r, err := h.client.R().
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

// AuthGRPC provides gRPC methods for auth operations.
type AuthGRPC struct {
	client pb.AuthServiceClient
}

// NewAuthGRPC constructs AuthGRPC with given gRPC client connection and creates the gRPC client.
func NewAuthGRPC(conn *grpc.ClientConn) *AuthGRPC {
	client := pb.NewAuthServiceClient(conn)
	return &AuthGRPC{client: client}
}

// Register calls the Register RPC method via gRPC.
func (g *AuthGRPC) Register(ctx context.Context, req *models.AuthRegisterRequest) (*models.AuthResponse, error) {
	pbReq := &pb.AuthRegisterRequest{
		Username:         req.Username,
		Password:         req.Password,
		ClientPubKeyFile: req.ClientPubKeyFile,
	}

	pbResp, err := g.client.Register(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("grpc register failed: %w", err)
	}

	return &models.AuthResponse{Token: pbResp.Token}, nil
}

// Login calls the Login RPC method via gRPC.
func (g *AuthGRPC) Login(ctx context.Context, req *models.AuthLoginRequest) (*models.AuthResponse, error) {
	pbReq := &pb.AuthLoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	pbResp, err := g.client.Login(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("grpc login failed: %w", err)
	}

	return &models.AuthResponse{Token: pbResp.Token}, nil
}
