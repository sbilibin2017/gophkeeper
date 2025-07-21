package text

import (
	"context"
	"io"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/text"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TextReadGRPCFacade implements read operations for text secrets over gRPC.
type TextReadGRPCFacade struct {
	client pb.TextReadServiceClient
}

// NewTextReadGRPCFacade creates a new TextReadGRPCFacade.
func NewTextReadGRPCFacade(client pb.TextReadServiceClient) *TextReadGRPCFacade {
	return &TextReadGRPCFacade{client: client}
}

// Get retrieves a text secret by secret name via gRPC.
func (g *TextReadGRPCFacade) Get(ctx context.Context, secretName string) (*models.TextDB, error) {
	req := &pb.TextGetRequest{SecretName: secretName}
	resp, err := g.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return &models.TextDB{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Content:     resp.Content,
		Meta:        resp.Meta,
		UpdatedAt:   resp.UpdatedAt,
	}, nil
}

// List retrieves all text secrets via gRPC streaming.
func (g *TextReadGRPCFacade) List(ctx context.Context) ([]models.TextDB, error) {
	stream, err := g.client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []models.TextDB
	for {
		textSecret, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, models.TextDB{
			SecretName:  textSecret.SecretName,
			SecretOwner: textSecret.SecretOwner,
			Content:     textSecret.Content,
			Meta:        textSecret.Meta,
			UpdatedAt:   textSecret.UpdatedAt,
		})
	}
	return results, nil
}
