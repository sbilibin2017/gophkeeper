package services

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- LoginHTTP with functional options ---

type LoginHTTPOpt func(*loginHTTPConfig)

type loginHTTPConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   *resty.Client
}

func WithLoginHTTPEncoders(enc []func([]byte) ([]byte, error)) LoginHTTPOpt {
	return func(c *loginHTTPConfig) {
		c.encoders = enc
	}
}

func WithLoginHTTPClient(client *resty.Client) LoginHTTPOpt {
	return func(c *loginHTTPConfig) {
		c.client = client
	}
}

func LoginHTTP(ctx context.Context, user *models.User, opts ...LoginHTTPOpt) error {
	config := &loginHTTPConfig{}
	for _, opt := range opts {
		opt(config)
	}

	encode := func(data string) (string, error) {
		b := []byte(data)
		var err error
		for _, enc := range config.encoders {
			b, err = enc(b)
			if err != nil {
				return "", err
			}
		}
		return base64.StdEncoding.EncodeToString(b), nil
	}

	encodedUsername, err := encode(user.Username)
	if err != nil {
		return fmt.Errorf("encoding username failed: %w", err)
	}

	encodedPassword, err := encode(user.Password)
	if err != nil {
		return fmt.Errorf("encoding password failed: %w", err)
	}

	data := map[string]string{
		"username": encodedUsername,
		"password": encodedPassword,
	}

	resp, err := config.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post("/login")
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.String())
	}

	return nil
}

// --- LoginGRPC with functional options ---

type LoginGRPCOpt func(*loginGRPCConfig)

type loginGRPCConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   pb.LoginServiceClient
}

func WithLoginGRPCEncoders(enc []func([]byte) ([]byte, error)) LoginGRPCOpt {
	return func(c *loginGRPCConfig) {
		c.encoders = enc
	}
}

func WithLoginGRPCClient(client pb.LoginServiceClient) LoginGRPCOpt {
	return func(c *loginGRPCConfig) {
		c.client = client
	}
}

func LoginGRPC(ctx context.Context, user *models.User, opts ...LoginGRPCOpt) error {
	config := &loginGRPCConfig{}
	for _, opt := range opts {
		opt(config)
	}

	encode := func(data string) (string, error) {
		b := []byte(data)
		var err error
		for _, enc := range config.encoders {
			b, err = enc(b)
			if err != nil {
				return "", err
			}
		}
		return base64.StdEncoding.EncodeToString(b), nil
	}

	encodedUsername, err := encode(user.Username)
	if err != nil {
		return fmt.Errorf("encoding username failed: %w", err)
	}

	encodedPassword, err := encode(user.Password)
	if err != nil {
		return fmt.Errorf("encoding password failed: %w", err)
	}

	req := &pb.LoginRequest{
		Username: encodedUsername,
		Password: encodedPassword,
	}

	resp, err := config.client.Login(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC request failed: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("login failed: %s", resp.Error)
	}

	return nil
}
