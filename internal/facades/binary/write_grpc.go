package binary

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/binary"
)

// GRPCBinaryWriteFacade implements write operations for binary secrets over gRPC.
type BinaryWriteGRPCFacade struct {
	client pb.BinaryWriteServiceClient
}

// NewGRPCBinaryWriteFacade creates a new GRPCBinaryWriteFacade.
func NewBinaryWriteGRPCFacade(client pb.BinaryWriteServiceClient) *BinaryWriteGRPCFacade {
	return &BinaryWriteGRPCFacade{client: client}
}

// Add calls the gRPC Add method to add a new binary secret.
// Converts the optional Meta pointer to a string before sending.
// Returns an error if the gRPC call fails.
func (g *BinaryWriteGRPCFacade) Add(ctx context.Context, req *models.BinaryAddRequest) error {
	var meta string
	if req.Meta != nil {
		meta = *req.Meta
	}

	grpcReq := &pb.BinaryAddRequest{
		SecretName: req.SecretName,
		Data:       req.Data,
		Meta:       meta,
	}
	_, err := g.client.Add(ctx, grpcReq)
	return err
}

// Delete calls the gRPC Delete method to delete a binary secret by secret name.
// Returns an error if the call fails.
func (g *BinaryWriteGRPCFacade) Delete(ctx context.Context, secretName string) error {
	req := &pb.BinaryDeleteRequest{SecretName: secretName}
	_, err := g.client.Delete(ctx, req)
	return err
}
