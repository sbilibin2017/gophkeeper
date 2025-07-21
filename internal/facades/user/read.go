package user

import (
	"context"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/user"
	"google.golang.org/protobuf/types/known/emptypb"
)

// UserReadHTTPFacade implements read operations for user secrets over HTTP.
type UserReadHTTPFacade struct {
	client *resty.Client
}

// NewUserReadHTTPFacade creates a new UserReadHTTPFacade.
func NewUserReadHTTPFacade(client *resty.Client) *UserReadHTTPFacade {
	return &UserReadHTTPFacade{client: client}
}

// Get retrieves a user secret by secret name via HTTP GET.
func (h *UserReadHTTPFacade) Get(ctx context.Context, secretName string) (*models.UserDB, error) {
	var respModel models.UserDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetQueryParam("secret_name", secretName).
		SetResult(&respModel).
		Get("/user/get")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get user secret: %s", resp.Status())
	}
	return &respModel, nil
}

// List retrieves all user secrets via HTTP GET.
func (h *UserReadHTTPFacade) List(ctx context.Context) ([]models.UserDB, error) {
	var respModel []models.UserDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetResult(&respModel).
		Get("/user/list")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to list user secrets: %s", resp.Status())
	}
	return respModel, nil
}

// UserReadGRPCFacade implements read operations for user secrets over gRPC.
type UserReadGRPCFacade struct {
	client pb.UserReadServiceClient
}

// NewUserReadGRPCFacade creates a new UserReadGRPCFacade.
func NewUserReadGRPCFacade(client pb.UserReadServiceClient) *UserReadGRPCFacade {
	return &UserReadGRPCFacade{client: client}
}

// Get retrieves a user secret by secret name via gRPC.
func (g *UserReadGRPCFacade) Get(ctx context.Context, secretName string) (*models.UserDB, error) {
	req := &pb.UserGetRequest{SecretName: secretName}
	resp, err := g.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	var metaPtr *string
	if resp.Meta != "" {
		metaPtr = &resp.Meta
	}

	return &models.UserDB{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Username:    resp.Username,
		Password:    resp.Password,
		Meta:        metaPtr,
		UpdatedAt:   resp.UpdatedAt,
	}, nil
}

// List retrieves all user secrets via gRPC streaming.
func (g *UserReadGRPCFacade) List(ctx context.Context) ([]models.UserDB, error) {
	stream, err := g.client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []models.UserDB
	for {
		userSecret, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		var metaPtr *string
		if userSecret.Meta != "" {
			metaPtr = &userSecret.Meta
		}

		results = append(results, models.UserDB{
			SecretName:  userSecret.SecretName,
			SecretOwner: userSecret.SecretOwner,
			Username:    userSecret.Username,
			Password:    userSecret.Password,
			Meta:        metaPtr,
			UpdatedAt:   userSecret.UpdatedAt,
		})
	}
	return results, nil
}
