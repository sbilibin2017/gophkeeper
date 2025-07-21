package text

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
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
