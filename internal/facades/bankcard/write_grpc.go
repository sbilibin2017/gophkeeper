package bankcard

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/bankcard"
)

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
