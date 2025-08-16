package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// DeviceHTTPFacade предоставляет методы для работы с устройствами через HTTP API.
type DeviceHTTPFacade struct {
	client *resty.Client
}

// NewDeviceHTTPFacade создаёт новый экземпляр DeviceHTTPFacade с указанным HTTP клиентом.
func NewDeviceHTTPFacade(client *resty.Client) *DeviceHTTPFacade {
	return &DeviceHTTPFacade{client: client}
}

// Get возвращает информацию о текущем устройстве пользователя.
func (h *DeviceHTTPFacade) Get(
	ctx context.Context,
	userID, deviceID string,
) (*models.DeviceDB, error) {
	var secret models.DeviceDB

	resp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(userID).
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
