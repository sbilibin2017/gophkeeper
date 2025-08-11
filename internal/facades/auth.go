package facades

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AuthHTTPFacade provides HTTP-based authentication methods.
type AuthHTTPFacade struct {
	client *resty.Client
}

// NewAuthHTTPFacade creates a new AuthHTTPFacade with the given Resty client.
func NewAuthHTTPFacade(client *resty.Client) *AuthHTTPFacade {
	return &AuthHTTPFacade{client: client}
}

// Register registers a new user via HTTP API.
// Returns a JWT token if successful.
func (a *AuthHTTPFacade) Register(ctx context.Context, username string, password string) (string, error) {
	body := map[string]string{
		"username": username,
		"password": password,
	}
	resp, err := a.client.R().
		SetContext(ctx).
		SetBody(body).
		Post("/register")
	if err != nil {
		return "", fmt.Errorf("register request failed: %w", err)
	}
	if resp.IsError() {
		return "", fmt.Errorf("register request returned error: %s", resp.Status())
	}

	authHeader := resp.Header().Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing in response")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("authorization header format invalid")
	}

	token := authHeader[len(bearerPrefix):]

	return token, nil
}

// Login authenticates a user via HTTP API.
// Returns a JWT token if successful.
func (a *AuthHTTPFacade) Login(ctx context.Context, username string, password string) (string, error) {
	body := map[string]string{
		"username": username,
		"password": password,
	}
	resp, err := a.client.R().
		SetContext(ctx).
		SetBody(body).
		Post("/login")
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	if resp.IsError() {
		return "", fmt.Errorf("login request returned error: %s", resp.Status())
	}

	authHeader := resp.Header().Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing in response")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("authorization header format invalid")
	}

	token := authHeader[len(bearerPrefix):]

	return token, nil
}

// GetUsername fetches the username associated with the provided token via HTTP API.
func (a *AuthHTTPFacade) GetUsername(ctx context.Context, token string) (string, error) {
	resp, err := a.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		Get("/username")
	if err != nil {
		return "", fmt.Errorf("get user request failed: %w", err)
	}
	if resp.IsError() {
		return "", fmt.Errorf("get user request returned error: %s", resp.Status())
	}

	// Local struct to unmarshal the expected JSON response
	var result struct {
		Username string `json:"username"`
	}

	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal user response: %w", err)
	}

	return result.Username, nil
}

// AuthGRPCFacade provides gRPC-based authentication methods.
type AuthGRPCFacade struct {
	client pb.AuthServiceClient
}

// NewAuthGRPCFacade creates a new AuthGRPCFacade using the given gRPC client connection.
func NewAuthGRPCFacade(client *grpc.ClientConn) *AuthGRPCFacade {
	return &AuthGRPCFacade{
		client: pb.NewAuthServiceClient(client),
	}
}

// Register registers a new user via gRPC and returns a JWT token if successful.
func (a *AuthGRPCFacade) Register(ctx context.Context, username string, password string) (string, error) {
	resp, err := a.client.Register(ctx, &pb.AuthRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}

// Login authenticates a user via gRPC and returns a JWT token if successful.
func (a *AuthGRPCFacade) Login(ctx context.Context, username string, password string) (string, error) {
	resp, err := a.client.Login(ctx, &pb.AuthRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}

// GetUsername fetches the username associated with the provided token via gRPC.
func (a *AuthGRPCFacade) GetUsername(ctx context.Context, token string) (string, error) {
	md := metadata.Pairs("authorization", "Bearer "+token)
	mdCtx := metadata.NewOutgoingContext(ctx, md)

	resp, err := a.client.GetUsername(mdCtx, &emptypb.Empty{})
	if err != nil {
		return "", err
	}

	return resp.Username, nil
}
