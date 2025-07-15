package clients

import (
	"time"

	"github.com/go-resty/resty/v2"
)

// NewHTTPClient создаёт новый HTTP клиент Resty с указанным базовым URL.
func NewHTTPClient(baseURL string) *resty.Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetRetryCount(3).                         // Повторять запросы до 3 раз при ошибках
		SetRetryWaitTime(500 * time.Millisecond). // Минимальная задержка между попытками
		SetRetryMaxWaitTime(2 * time.Second)      // Максимальная задержка между попытками
	return client
}
