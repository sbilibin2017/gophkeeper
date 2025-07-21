package binary

import (
	"context"
	"io"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/binary"
	"google.golang.org/protobuf/types/known/emptypb"
)

// BinaryReadGRPCFacade implements read operations for binary secrets over gRPC.
type BinaryReadGRPCFacade struct {
	client pb.BinaryReadServiceClient
}

// NewBinaryReadGRPCFacade creates a new BinaryReadGRPCFacade.
func NewBinaryReadGRPCFacade(client pb.BinaryReadServiceClient) *BinaryReadGRPCFacade {
	return &BinaryReadGRPCFacade{client: client}
}

// Get retrieves a binary secret by secret name via gRPC.
func (g *BinaryReadGRPCFacade) Get(ctx context.Context, secretName string) (*models.BinaryDB, error) {
	req := &pb.BinaryGetRequest{SecretName: secretName}
	resp, err := g.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return &models.BinaryDB{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Data:        resp.Data,
		Meta:        &resp.Meta,
		UpdatedAt:   resp.UpdatedAt,
	}, nil
}

// List retrieves all binary secrets via gRPC streaming.
func (g *BinaryReadGRPCFacade) List(ctx context.Context) ([]models.BinaryDB, error) {
	stream, err := g.client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []models.BinaryDB
	for {
		binarySecret, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, models.BinaryDB{
			SecretName:  binarySecret.SecretName,
			SecretOwner: binarySecret.SecretOwner,
			Data:        binarySecret.Data,
			Meta:        &binarySecret.Meta,
			UpdatedAt:   binarySecret.UpdatedAt,
		})
	}
	return results, nil
}
