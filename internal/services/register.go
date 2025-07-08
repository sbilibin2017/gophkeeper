package services

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// --- RegisterHTTP with functional options (without ctx) ---

type RegisterHTTPOpt func(*registerHTTPConfig)

type registerHTTPConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   *resty.Client
}

func WithRegisterHTTPEncoders(enc []func([]byte) ([]byte, error)) RegisterHTTPOpt {
	return func(c *registerHTTPConfig) {
		c.encoders = enc
	}
}

func WithRegisterHTTPClient(client *resty.Client) RegisterHTTPOpt {
	return func(c *registerHTTPConfig) {
		c.client = client
	}
}

func RegisterHTTP(ctx context.Context, user *models.User, opts ...RegisterHTTPOpt) error {
	config := &registerHTTPConfig{}
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
		Post("/register")
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.String())
	}

	return nil
}

// --- RegisterGRPC with functional options (without ctx) ---

type RegisterGRPCOpt func(*registerGRPCConfig)

type registerGRPCConfig struct {
	encoders []func([]byte) ([]byte, error)
	client   pb.RegisterServiceClient
}

func WithRegisterGRPCEncoders(enc []func([]byte) ([]byte, error)) RegisterGRPCOpt {
	return func(c *registerGRPCConfig) {
		c.encoders = enc
	}
}

func WithRegisterGRPCClient(client pb.RegisterServiceClient) RegisterGRPCOpt {
	return func(c *registerGRPCConfig) {
		c.client = client
	}
}

func RegisterGRPC(ctx context.Context, user *models.User, opts ...RegisterGRPCOpt) error {
	config := &registerGRPCConfig{}
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

	req := &pb.RegisterRequest{
		Username: encodedUsername,
		Password: encodedPassword,
	}

	resp, err := config.client.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC request failed: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("registration failed: %s", resp.Error)
	}

	return nil
}
