package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// LoginHTTP sends user credentials to the /login HTTP endpoint and returns a JWT token.
func LoginHTTP(ctx context.Context, client *resty.Client, cred *models.Credentials) (string, error) {
	const op = "services.LoginHTTP"

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(cred).
		Post("/login")
	if err != nil {
		return "", fmt.Errorf("%s: request failed: %w", op, err)
	}
	if resp.IsError() {
		return "", fmt.Errorf("%s: client internal error, status code: %d", op, resp.StatusCode())
	}

	var respData struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(resp.Body(), &respData); err != nil {
		return "", fmt.Errorf("%s: failed to unmarshal response: %w", op, err)
	}

	return respData.Token, nil
}

// LoginGRPC calls the gRPC Login method and returns a JWT token.
func LoginGRPC(ctx context.Context, client pb.LoginServiceClient, cred *models.Credentials) (string, error) {
	const op = "services.LoginGRPC"

	req := &pb.Credentials{
		Username: cred.Username,
		Password: cred.Password,
	}

	resp, err := client.Login(ctx, req)
	if err != nil {
		return "", fmt.Errorf("%s: login RPC call failed: %w", op, err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("%s: login failed: %s", op, resp.Error)
	}

	return resp.Token, nil
}

// RegisterHTTP sends user credentials to the /register HTTP endpoint and returns a JWT token.
func RegisterHTTP(ctx context.Context, client *resty.Client, cred *models.Credentials) (string, error) {
	const op = "services.RegisterHTTP"

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(cred).
		Post("/register")
	if err != nil {
		return "", fmt.Errorf("%s: request failed: %w", op, err)
	}
	if resp.IsError() {
		return "", fmt.Errorf("%s: server returned error, status code %d: %s", op, resp.StatusCode(), resp.String())
	}

	var respData struct {
		Token string `json:"token"`
	}

	if err := json.Unmarshal(resp.Body(), &respData); err != nil {
		return "", fmt.Errorf("%s: failed to unmarshal response: %w", op, err)
	}

	if respData.Token == "" {
		return "", fmt.Errorf("%s: token not found in response", op)
	}

	return respData.Token, nil
}

// RegisterGRPC calls the gRPC Register method and returns a JWT token.
func RegisterGRPC(ctx context.Context, client pb.RegisterServiceClient, cred *models.Credentials) (string, error) {
	const op = "services.RegisterGRPC"

	req := &pb.Credentials{
		Username: cred.Username,
		Password: cred.Password,
	}

	resp, err := client.Register(ctx, req)
	if err != nil {
		return "", fmt.Errorf("%s: register RPC call failed: %w", op, err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("%s: registration failed: %s", op, resp.Error)
	}

	return resp.Token, nil
}
