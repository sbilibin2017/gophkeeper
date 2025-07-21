package text

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/text"
)

// TextWriteGRPCFacade implements write operations for text secrets over gRPC.
type TextWriteGRPCFacade struct {
	client pb.TextWriteServiceClient
}

// NewTextWriteGRPCFacade creates a new TextWriteGRPCFacade.
func NewTextWriteGRPCFacade(client pb.TextWriteServiceClient) *TextWriteGRPCFacade {
	return &TextWriteGRPCFacade{client: client}
}

// Add calls the gRPC Add method to add a new text secret.
// Converts the optional Meta pointer to a string before sending.
// Returns an error if the gRPC call fails.
func (g *TextWriteGRPCFacade) Add(ctx context.Context, req *models.TextAddRequest) error {
	var meta string
	if req.Meta != nil {
		meta = *req.Meta
	}

	grpcReq := &pb.TextAddRequest{
		SecretName: req.SecretName,
		Content:    req.Content,
		Meta:       meta,
	}
	_, err := g.client.Add(ctx, grpcReq)
	return err
}

// Delete calls the gRPC Delete method to delete a text secret by secret name.
// Returns an error if the call fails.
func (g *TextWriteGRPCFacade) Delete(ctx context.Context, secretName string) error {
	req := &pb.TextDeleteRequest{SecretName: secretName}
	_, err := g.client.Delete(ctx, req)
	return err
}
