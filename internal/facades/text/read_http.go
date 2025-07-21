package text

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// TextReadHTTPFacade implements read operations for text secrets over HTTP.
type TextReadHTTPFacade struct {
	client *resty.Client
}

// NewTextReadHTTPFacade creates a new TextReadHTTPFacade.
func NewTextReadHTTPFacade(client *resty.Client) *TextReadHTTPFacade {
	return &TextReadHTTPFacade{client: client}
}

// Get retrieves a text secret by secret name via HTTP GET.
func (h *TextReadHTTPFacade) Get(ctx context.Context, secretName string) (*models.TextDB, error) {
	var respModel models.TextDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetQueryParam("secret_name", secretName).
		SetResult(&respModel).
		Get("/text/get")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get text secret: %s", resp.Status())
	}
	return &respModel, nil
}

// List retrieves all text secrets via HTTP GET.
func (h *TextReadHTTPFacade) List(ctx context.Context) ([]models.TextDB, error) {
	var respModel []models.TextDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetResult(&respModel).
		Get("/text/list")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to list text secrets: %s", resp.Status())
	}
	return respModel, nil
}
