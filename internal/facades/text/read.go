package text

import (
	"context"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/text"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TextReadHTTPFacade implements read operations for text secrets over HTTP.
type TextReadHTTPFacade struct {
	client *resty.Client
}

// NewTextReadHTTPFacade creates a new TextReadHTTPFacade.
func NewTextReadHTTPFacade(client *resty.Client) *TextReadHTTPFacade {
	return &TextReadHTTPFacade{client: client}
}

// Get retrieves a text secret by secret name via HTTP GET.
func (h *TextReadHTTPFacade) Get(ctx context.Context, secretName string) (*models.TextDB, error) {
	var respModel models.TextDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetQueryParam("secret_name", secretName).
		SetResult(&respModel).
		Get("/text/get")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get text secret: %s", resp.Status())
	}
	return &respModel, nil
}

// List retrieves all text secrets via HTTP GET.
func (h *TextReadHTTPFacade) List(ctx context.Context) ([]models.TextDB, error) {
	var respModel []models.TextDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetResult(&respModel).
		Get("/text/list")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to list text secrets: %s", resp.Status())
	}
	return respModel, nil
}

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
