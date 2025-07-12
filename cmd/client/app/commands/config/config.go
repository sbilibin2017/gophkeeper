package config

import (
	"errors"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
)

// NewConfig создает конфигурацию клиента.
// Поддерживает HTTP и gRPC протоколы.
// Возвращает указатель на ClientConfig или ошибку при неверном формате URL или проблемах с созданием.
func NewConfig(serverURL string) (*configs.ClientConfig, error) {
	var opts []configs.ClientConfigOpt

	switch {
	case strings.HasPrefix(serverURL, "http://"), strings.HasPrefix(serverURL, "https://"):
		opts = append(opts, configs.WithHTTPClient(serverURL))
	case strings.HasPrefix(serverURL, "grpc://"):
		opts = append(opts, configs.WithGRPCClient(serverURL))
	default:
		return nil, errors.New("неверный формат URL сервера")
	}

	config, err := configs.NewClientConfig(opts...)
	if err != nil {
		return nil, errors.New("не удалось создать конфигурацию клиента")
	}

	return config, nil
}
