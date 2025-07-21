package bankcard

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
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
