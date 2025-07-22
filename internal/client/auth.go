package client

import (
	"context"
	"errors"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RegisterHTTP calls the HTTP /register endpoint with JSON body.
func RegisterHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	resp := &models.AuthResponse{}
	r, err := client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(resp).
		Post("/register")
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(r.Status())
	}
	return resp, nil
}

// RegisterGRPC calls the Register RPC method.
func RegisterGRPC(
	ctx context.Context,
	client pb.AuthServiceClient,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	grpcReq := &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}
	grpcResp, err := client.Register(ctx, grpcReq)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{Token: grpcResp.Token}, nil
}

// LoginHTTP calls the HTTP /login endpoint with JSON body.
func LoginHTTP(
	ctx context.Context,
	client *resty.Client,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	resp := &models.AuthResponse{}
	r, err := client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(resp).
		Post("/login")
	if err != nil {
		return nil, err
	}
	if r.IsError() {
		return nil, errors.New(r.Status())
	}
	return resp, nil
}

// LoginGRPC calls the Login RPC method.
func LoginGRPC(
	ctx context.Context,
	client pb.AuthServiceClient,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	grpcReq := &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	}
	grpcResp, err := client.Login(ctx, grpcReq)
	if err != nil {
		return nil, err
	}
	return &models.AuthResponse{Token: grpcResp.Token}, nil
}

// LogoutHTTP calls the HTTP /logout endpoint.
// No body needed, returns empty response.
func LogoutHTTP(
	ctx context.Context,
	client *resty.Client,
) error {
	r, err := client.R().
		SetContext(ctx).
		Post("/logout")
	if err != nil {
		return err
	}
	if r.IsError() {
		return errors.New(r.Status())
	}
	return nil
}

// LogoutGRPC calls the Logout RPC method.
func LogoutGRPC(
	ctx context.Context,
	client pb.AuthServiceClient,
) error {
	_, err := client.Logout(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	return nil
}
