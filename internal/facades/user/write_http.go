package user

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// UserWriteHTTPFacade implements write operations for user secrets over HTTP.
type UserWriteHTTPFacade struct {
	client *resty.Client
}

// NewUserWriteHTTPFacade creates a new UserWriteHTTPFacade.
func NewUserWriteHTTPFacade(client *resty.Client) *UserWriteHTTPFacade {
	return &UserWriteHTTPFacade{client: client}
}

// Add sends an HTTP POST request to add a new user secret.
// Returns an error if the request fails or the server responds with a non-200 status.
func (h *UserWriteHTTPFacade) Add(ctx context.Context, req *models.UserAddRequest) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/user/add")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to add user secret: %s", resp.Status())
	}
	return nil
}

// Delete sends an HTTP POST request to delete a user secret by secret name.
// Returns an error if the request fails or the server responds with a non-200 status.
func (h *UserWriteHTTPFacade) Delete(ctx context.Context, secretName string) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetBody(map[string]string{"secret_name": secretName}).
		Post("/user/delete")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to delete user secret: %s", resp.Status())
	}
	return nil
}
