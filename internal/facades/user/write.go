package user

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/user"
)

// UserWriteHTTPFacade implements write operations for user secrets over HTTP.
type UserWriteHTTPFacade struct {
	client *resty.Client
}

// NewUserWriteHTTPFacade creates a new UserWriteHTTPFacade.
func NewUserWriteHTTPFacade(client *resty.Client) *UserWriteHTTPFacade {
	return &UserWriteHTTPFacade{client: client}
}

// Add sends an HTTP POST request to add a new user secret.
// Returns an error if the request fails or the server responds with a non-200 status.
func (h *UserWriteHTTPFacade) Add(ctx context.Context, req *models.UserAddRequest) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/user/add")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to add user secret: %s", resp.Status())
	}
	return nil
}

// Delete sends an HTTP POST request to delete a user secret by secret name.
// Returns an error if the request fails or the server responds with a non-200 status.
func (h *UserWriteHTTPFacade) Delete(ctx context.Context, secretName string) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"secret_name": secretName}).
		Post("/user/delete")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to delete user secret: %s", resp.Status())
	}
	return nil
}

// UserWriteGRPCFacade implements write operations for user secrets over gRPC.
type UserWriteGRPCFacade struct {
	client pb.UserWriteServiceClient
}

// NewUserWriteGRPCFacade creates a new UserWriteGRPCFacade.
func NewUserWriteGRPCFacade(client pb.UserWriteServiceClient) *UserWriteGRPCFacade {
	return &UserWriteGRPCFacade{client: client}
}

// Add calls the gRPC Add method to add a new user secret.
// Converts the optional Meta pointer to a string before sending.
// Returns an error if the gRPC call fails.
func (g *UserWriteGRPCFacade) Add(ctx context.Context, req *models.UserAddRequest) error {
	var meta string
	if req.Meta != nil {
		meta = *req.Meta
	}

	grpcReq := &pb.UserAddRequest{
		SecretName: req.SecretName,
		Username:   req.Username,
		Password:   req.Password,
		Meta:       meta,
	}
	_, err := g.client.Add(ctx, grpcReq)
	return err
}

// Delete calls the gRPC Delete method to delete a user secret by secret name.
// Returns an error if the call fails.
func (g *UserWriteGRPCFacade) Delete(ctx context.Context, secretName string) error {
	req := &pb.UserDeleteRequest{SecretName: secretName}
	_, err := g.client.Delete(ctx, req)
	return err
}
