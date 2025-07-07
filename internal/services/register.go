package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// private options struct for RegisterHTTP
type registerHTTPOptions struct {
	client   *resty.Client
	encoders []func(data []byte) ([]byte, error)
}

type RegisterHTTPOption func(*registerHTTPOptions)

func WithRegisterHTTPClient(client *resty.Client) RegisterHTTPOption {
	return func(o *registerHTTPOptions) {
		o.client = client
	}
}

// Wrap HMAC encoder (which returns []byte) to match signature ([]byte, error)
func WithRegisterHMACEncoder(enc func([]byte) []byte) RegisterHTTPOption {
	return func(o *registerHTTPOptions) {
		o.encoders = append(o.encoders, func(data []byte) ([]byte, error) {
			return enc(data), nil
		})
	}
}

func WithRegisterRSAEncoder(enc func([]byte) ([]byte, error)) RegisterHTTPOption {
	return func(o *registerHTTPOptions) {
		o.encoders = append(o.encoders, enc)
	}
}

// RegisterHTTP registers a user via HTTP API.
// ctx, username, password are required parameters;
// optional params are set via functional options.
func RegisterHTTP(
	ctx context.Context,
	username, password string,
	opts ...RegisterHTTPOption,
) error {
	options := &registerHTTPOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Encode username
	encodedUsername := []byte(username)
	var err error
	for _, enc := range options.encoders {
		encodedUsername, err = enc(encodedUsername)
		if err != nil {
			return fmt.Errorf("failed to encode username: %w", err)
		}
	}
	encodedUsernameHex := hex.EncodeToString(encodedUsername)

	// Encode password
	encodedPassword := []byte(password)
	for _, enc := range options.encoders {
		encodedPassword, err = enc(encodedPassword)
		if err != nil {
			return fmt.Errorf("failed to encode password: %w", err)
		}
	}
	encodedPasswordHex := hex.EncodeToString(encodedPassword)

	resp, err := options.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"username": encodedUsernameHex,
			"password": encodedPasswordHex,
		}).
		Post("/register")

	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("registration failed: %s", resp.String())
	}

	return nil
}

// ------------------------------------------------------------

// private options struct for RegisterGRPC
type registerGRPCOptions struct {
	client   pb.RegisterServiceClient
	encoders []func(data []byte) ([]byte, error)
}

type RegisterGRPCOption func(*registerGRPCOptions)

func WithRegisterGRPCClient(client pb.RegisterServiceClient) RegisterGRPCOption {
	return func(o *registerGRPCOptions) {
		o.client = client
	}
}

func WithRegisterGRPCHMACEncoder(enc func([]byte) []byte) RegisterGRPCOption {
	return func(o *registerGRPCOptions) {
		o.encoders = append(o.encoders, func(data []byte) ([]byte, error) {
			return enc(data), nil
		})
	}
}

func WithRegisterGRPCRSAEncoder(enc func([]byte) ([]byte, error)) RegisterGRPCOption {
	return func(o *registerGRPCOptions) {
		o.encoders = append(o.encoders, enc)
	}
}

// RegisterGRPC registers a user via gRPC.
// ctx, username, password are required parameters;
// optional params are set via functional options.
func RegisterGRPC(
	ctx context.Context,
	username, password string,
	opts ...RegisterGRPCOption,
) error {
	options := &registerGRPCOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Encode username
	encodedUsername := []byte(username)
	var err error
	for _, enc := range options.encoders {
		encodedUsername, err = enc(encodedUsername)
		if err != nil {
			return fmt.Errorf("failed to encode username: %w", err)
		}
	}

	// Encode password
	encodedPassword := []byte(password)
	for _, enc := range options.encoders {
		encodedPassword, err = enc(encodedPassword)
		if err != nil {
			return fmt.Errorf("failed to encode password: %w", err)
		}
	}

	_, err = options.client.Register(ctx, &pb.RegisterRequest{
		Username: string(encodedUsername), // encode as base64/hex if needed
		Password: string(encodedPassword),
	})
	return err
}
