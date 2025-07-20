package bankcard

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/bankcard"
)

// HTTPClient обёртка для HTTP-клиента, работающего с банковскими картами.
type BankCardAddHTTPFacade struct {
	client *resty.Client
}

// NewHTTPClient создаёт новый HTTPClient.
func NewBankCardAddHTTPFacade(client *resty.Client) *BankCardAddHTTPFacade {
	return &BankCardAddHTTPFacade{client: client}
}

// Add отправляет запрос добавления банковской карты по HTTP.
func (h *BankCardAddHTTPFacade) Add(ctx context.Context, req *models.BankCardAddRequest) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/bankcard/add")
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("http add bank card failed: %s", resp.Status())
	}
	return nil
}

// GRPCClient обёртка для gRPC-клиента, работающего с банковскими картами.
type BankCardAddGRPCFacade struct {
	client pb.BankCardWriteServiceClient
}

// NewGRPCClient создаёт новый GRPCClient.
func NewBankCardAddGRPCFacade(client pb.BankCardWriteServiceClient) *BankCardAddGRPCFacade {
	return &BankCardAddGRPCFacade{client: client}
}

// Add отправляет запрос добавления банковской карты через gRPC.
func (g *BankCardAddGRPCFacade) Add(ctx context.Context, req *models.BankCardAddRequest) error {
	meta := ""
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
