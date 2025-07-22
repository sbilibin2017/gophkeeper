package text

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/text"
)

// TextWriteHTTPFacade implements write operations for text secrets over HTTP.
type TextWriteHTTPFacade struct {
	client *resty.Client
}

// NewTextWriteHTTPFacade creates a new TextWriteHTTPFacade.
func NewTextWriteHTTPFacade(client *resty.Client) *TextWriteHTTPFacade {
	return &TextWriteHTTPFacade{client: client}
}

// Add sends an HTTP POST request to add a new text secret.
// Returns an error if the request fails or the server responds with a non-200 status.
func (h *TextWriteHTTPFacade) Add(ctx context.Context, req *models.TextAddRequest) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/text/add")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to add text secret: %s", resp.Status())
	}
	return nil
}

// Delete sends an HTTP POST request to delete a text secret by secret name.
// Returns an error if the request fails or the server responds with a non-200 status.
func (h *TextWriteHTTPFacade) Delete(ctx context.Context, secretName string) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"secret_name": secretName}).
		Post("/text/delete")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to delete text secret: %s", resp.Status())
	}
	return nil
}

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
