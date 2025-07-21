package user

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/user"
)

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
