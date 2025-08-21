package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// DeviceHTTPFacade предоставляет методы для работы с устройствами через HTTP API.
type DeviceHTTPClient struct {
	client *resty.Client
}

// NewDeviceHTTPFacade создаёт новый экземпляр DeviceHTTPFacade с указанным HTTP клиентом.
func NewDeviceHTTPClient(client *resty.Client) *DeviceHTTPClient {
	return &DeviceHTTPClient{client: client}
}

// Get возвращает информацию о текущем устройстве пользователя.
func (h *DeviceHTTPClient) Get(
	ctx context.Context,
	token string,
) (*models.DeviceResponse, error) {
	var secret models.DeviceResponse

	resp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get("/get")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("http error: %s", resp.Status())
	}

	return &secret, nil
}
