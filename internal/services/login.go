package services

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/go-resty/resty/v2"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// private options struct for LoginHTTP
type loginHTTPOptions struct {
	client   *resty.Client
	encoders []func(data []byte) ([]byte, error)
}

type LoginHTTPOption func(*loginHTTPOptions)

func WithLoginHTTPClient(client *resty.Client) LoginHTTPOption {
	return func(o *loginHTTPOptions) {
		o.client = client
	}
}

func WithLoginHMACEncoder(enc func([]byte) []byte) LoginHTTPOption {
	return func(o *loginHTTPOptions) {
		o.encoders = append(o.encoders, func(data []byte) ([]byte, error) {
			return enc(data), nil
		})
	}
}

func WithLoginRSAEncoder(enc func([]byte) ([]byte, error)) LoginHTTPOption {
	return func(o *loginHTTPOptions) {
		o.encoders = append(o.encoders, enc)
	}
}

func LoginHTTP(
	ctx context.Context,
	username, password string,
	opts ...LoginHTTPOption,
) error {
	options := &loginHTTPOptions{}
	for _, opt := range opts {
		opt(options)
	}

	encodedUsername := []byte(username)
	var err error
	for _, enc := range options.encoders {
		encodedUsername, err = enc(encodedUsername)
		if err != nil {
			return fmt.Errorf("failed to encode username: %w", err)
		}
	}
	encodedUsernameHex := hex.EncodeToString(encodedUsername)

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
		Post("/login")

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("login failed: %s", resp.String())
	}

	return nil
}

// ------------------------------------------------------------

type loginGRPCOptions struct {
	client   pb.LoginServiceClient
	encoders []func(data []byte) ([]byte, error)
}

type LoginGRPCOption func(*loginGRPCOptions)

func WithLoginGRPCClient(client pb.LoginServiceClient) LoginGRPCOption {
	return func(o *loginGRPCOptions) {
		o.client = client
	}
}

func WithLoginGRPCHMACEncoder(enc func([]byte) []byte) LoginGRPCOption {
	return func(o *loginGRPCOptions) {
		o.encoders = append(o.encoders, func(data []byte) ([]byte, error) {
			return enc(data), nil
		})
	}
}

func WithLoginGRPCRSAEncoder(enc func([]byte) ([]byte, error)) LoginGRPCOption {
	return func(o *loginGRPCOptions) {
		o.encoders = append(o.encoders, enc)
	}
}

func LoginGRPC(
	ctx context.Context,
	username, password string,
	opts ...LoginGRPCOption,
) error {
	options := &loginGRPCOptions{}
	for _, opt := range opts {
		opt(options)
	}

	encodedUsername := []byte(username)
	var err error
	for _, enc := range options.encoders {
		encodedUsername, err = enc(encodedUsername)
		if err != nil {
			return fmt.Errorf("failed to encode username: %w", err)
		}
	}

	encodedPassword := []byte(password)
	for _, enc := range options.encoders {
		encodedPassword, err = enc(encodedPassword)
		if err != nil {
			return fmt.Errorf("failed to encode password: %w", err)
		}
	}

	encodedPasswordHex := hex.EncodeToString(encodedPassword)

	_, err = options.client.Login(ctx, &pb.LoginRequest{
		Username: string(encodedUsername),
		Password: encodedPasswordHex,
	})
	return err
}
