package bankcard

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
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
