package user

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// UserReadHTTPFacade implements read operations for user secrets over HTTP.
type UserReadHTTPFacade struct {
	client *resty.Client
}

// NewUserReadHTTPFacade creates a new UserReadHTTPFacade.
func NewUserReadHTTPFacade(client *resty.Client) *UserReadHTTPFacade {
	return &UserReadHTTPFacade{client: client}
}

// Get retrieves a user secret by secret name via HTTP GET.
func (h *UserReadHTTPFacade) Get(ctx context.Context, secretName string) (*models.UserDB, error) {
	var respModel models.UserDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetQueryParam("secret_name", secretName).
		SetResult(&respModel).
		Get("/user/get")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get user secret: %s", resp.Status())
	}
	return &respModel, nil
}

// List retrieves all user secrets via HTTP GET.
func (h *UserReadHTTPFacade) List(ctx context.Context) ([]models.UserDB, error) {
	var respModel []models.UserDB
	resp, err := h.client.R().
		SetContext(ctx).
		SetResult(&respModel).
		Get("/user/list")

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to list user secrets: %s", resp.Status())
	}
	return respModel, nil
}
