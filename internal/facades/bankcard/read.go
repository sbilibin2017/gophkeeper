package bankcard

import (
	"context"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/bankcard"
	"google.golang.org/protobuf/types/known/emptypb"
)

// BankCardReadHTTPFacade implements read operations for bank cards over HTTP.
type BankCardReadHTTPFacade struct {
	client *resty.Client
}

// NewHTTPBankCardReadService creates a new HTTPBankCardReadService.
func NewBankCardReadHTTPFacade(client *resty.Client) *BankCardReadHTTPFacade {
	return &BankCardReadHTTPFacade{client: client}
}

// Get retrieves a bank card by secret name via HTTP GET.
func (h *BankCardReadHTTPFacade) Get(ctx context.Context, secretName string) (*models.BankCardDB, error) {
	var respModel models.BankCardDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetQueryParam("secret_name", secretName).
		SetResult(&respModel).
		Get("/bankcard/get")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get bank card: %s", resp.Status())
	}
	return &respModel, nil
}

// List retrieves all bank cards via HTTP GET.
func (h *BankCardReadHTTPFacade) List(ctx context.Context) ([]models.BankCardDB, error) {
	var respModel []models.BankCardDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetResult(&respModel).
		Get("/bankcard/list")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to list bank cards: %s", resp.Status())
	}
	return respModel, nil
}

// GRPCBankCardReadService implements read operations for bank cards over gRPC.
type BankCardReadGRPCService struct {
	client pb.BankCardReadServiceClient
}

// NewGRPCBankCardReadService creates a new GRPCBankCardReadService.
func NewBankCardReadGRPCService(client pb.BankCardReadServiceClient) *BankCardReadGRPCService {
	return &BankCardReadGRPCService{client: client}
}

// Get retrieves a bank card by secret name via gRPC.
func (g *BankCardReadGRPCService) Get(ctx context.Context, secretName string) (*models.BankCardDB, error) {
	req := &pb.BankCardGetRequest{SecretName: secretName}
	resp, err := g.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return &models.BankCardDB{
		SecretName:  resp.SecretName,
		SecretOwner: resp.SecretOwner,
		Number:      resp.Number,
		Owner:       resp.Owner,
		Exp:         resp.Exp,
		CVV:         resp.Cvv,
		Meta:        resp.Meta,
		UpdatedAt:   resp.UpdatedAt,
	}, nil
}

// List retrieves all bank cards via gRPC streaming.
func (g *BankCardReadGRPCService) List(ctx context.Context) ([]models.BankCardDB, error) {
	stream, err := g.client.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var results []models.BankCardDB
	for {
		bankCard, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, models.BankCardDB{
			SecretName:  bankCard.SecretName,
			SecretOwner: bankCard.SecretOwner,
			Number:      bankCard.Number,
			Owner:       bankCard.Owner,
			Exp:         bankCard.Exp,
			CVV:         bankCard.Cvv,
			Meta:        bankCard.Meta,
			UpdatedAt:   bankCard.UpdatedAt,
		})
	}
	return results, nil
}
