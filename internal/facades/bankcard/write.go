package bankcard

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/bankcard"
)

// BankCardWriteHTTPFacade implements write operations for bank cards over HTTP.
type BankCardWriteHTTPFacade struct {
	client *resty.Client
}

// NewBankCardWriteHTTPFacade creates a new BankCardWriteHTTPFacade.
func NewBankCardWriteHTTPFacade(client *resty.Client) *BankCardWriteHTTPFacade {
	return &BankCardWriteHTTPFacade{client: client}
}

// Add sends an HTTP POST request to add a new bank card.
// Returns an error if the request fails or the server responds with a non-200 status.
func (h *BankCardWriteHTTPFacade) Add(ctx context.Context, req *models.BankCardAddRequest) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/bankcard/add")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to add bank card: %s", resp.Status())
	}
	return nil
}

// Delete sends an HTTP POST request to delete a bank card by secret name.
// Returns an error if the request fails or the server responds with a non-200 status.
func (h *BankCardWriteHTTPFacade) Delete(ctx context.Context, secretName string) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"secret_name": secretName}).
		Post("/bankcard/delete")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to delete bank card: %s", resp.Status())
	}
	return nil
}

// GRPCBankCardWriteService implements write operations for bank cards over gRPC.
type BankCardWriteGRPCFacade struct {
	client pb.BankCardWriteServiceClient
}

// NewGRPCBankCardWriteService creates a new GRPCBankCardWriteService.
func NewBankCardWriteGRPCFacade(client pb.BankCardWriteServiceClient) *BankCardWriteGRPCFacade {
	return &BankCardWriteGRPCFacade{client: client}
}

// Add calls the gRPC Add method to add a new bank card.
// Converts the optional Meta pointer to a string before sending.
// Returns an error if the gRPC call fails.
func (g *BankCardWriteGRPCFacade) Add(ctx context.Context, req *models.BankCardAddRequest) error {
	var meta string
	if req.Meta != nil {
		meta = *req.Meta
	}

	grpcReq := &pb.BankCardAddRequest{
		SecretName: req.SecretName,
		Number:     req.Number,
		Owner:      req.Owner,
		Exp:        req.Exp,
		Cvv:        req.CVV,
		Meta:       meta,
	}
	_, err := g.client.Add(ctx, grpcReq)
	return err
}

// Delete calls the gRPC Delete method to delete a bank card by secret name.
// Returns an error if the call fails.
func (g *BankCardWriteGRPCFacade) Delete(ctx context.Context, secretName string) error {
	req := &pb.BankCardDeleteRequest{SecretName: secretName}
	_, err := g.client.Delete(ctx, req)
	return err
}
