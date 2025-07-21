package binary

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// BinaryWriteHTTPFacade implements write operations for binary secrets over HTTP.
type BinaryWriteHTTPFacade struct {
	client *resty.Client
}

// NewBinaryWriteHTTPFacade creates a new BinaryWriteHTTPFacade.
func NewBinaryWriteHTTPFacade(client *resty.Client) *BinaryWriteHTTPFacade {
	return &BinaryWriteHTTPFacade{client: client}
}

// Add sends an HTTP POST request to add a new binary secret.
// Returns an error if the request fails or the server responds with a non-200 status.
func (h *BinaryWriteHTTPFacade) Add(ctx context.Context, req *models.BinaryAddRequest) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/binary/add")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to add binary secret: %s", resp.Status())
	}
	return nil
}

// Delete sends an HTTP POST request to delete a binary secret by secret name.
// Returns an error if the request fails or the server responds with a non-200 status.
func (h *BinaryWriteHTTPFacade) Delete(ctx context.Context, secretName string) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"secret_name": secretName}).
		Post("/binary/delete")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to delete binary secret: %s", resp.Status())
	}
	return nil
}
