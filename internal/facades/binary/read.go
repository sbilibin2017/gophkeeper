package binary

import (
	"context"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/binary"
	"google.golang.org/protobuf/types/known/emptypb"
)

// BinaryReadHTTPFacade implements read operations for binary secrets over HTTP.
type BinaryReadHTTPFacade struct {
	client *resty.Client
}

// NewBinaryReadHTTPFacade creates a new BinaryReadHTTPFacade.
func NewBinaryReadHTTPFacade(client *resty.Client) *BinaryReadHTTPFacade {
	return &BinaryReadHTTPFacade{client: client}
}

// Get retrieves a binary secret by secret name via HTTP GET.
func (h *BinaryReadHTTPFacade) Get(ctx context.Context, secretName string) (*models.BinaryDB, error) {
	var respModel models.BinaryDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetQueryParam("secret_name", secretName).
		SetResult(&respModel).
		Get("/binary/get")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get binary secret: %s", resp.Status())
	}
	return &respModel, nil
}

// List retrieves all binary secrets via HTTP GET.
func (h *BinaryReadHTTPFacade) List(ctx context.Context) ([]models.BinaryDB, error) {
	var respModel []models.BinaryDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetResult(&respModel).
		Get("/binary/list")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to list binary secrets: %s", resp.Status())
	}
	return respModel, nil
}

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
