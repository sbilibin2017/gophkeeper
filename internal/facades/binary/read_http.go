package binary

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
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
