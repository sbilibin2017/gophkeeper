package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClient(t *testing.T) {
	t.Run("valid URL", func(t *testing.T) {
		client, err := NewHTTPClient("http://localhost:8080")
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "http://localhost:8080", client.BaseURL)
	})

	t.Run("invalid URL", func(t *testing.T) {
		client, err := NewHTTPClient("://bad-url")
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestNewGRPCClient(t *testing.T) {
	t.Run("valid grpc URL", func(t *testing.T) {
		// Пример валидного URL с host:port
		client, err := NewGRPCClient("dns:///localhost:50051")
		assert.NoError(t, err)
		assert.NotNil(t, client)

		// Закрываем соединение после теста
		_ = client.Close()
	})

	t.Run("invalid grpc URL", func(t *testing.T) {
		client, err := NewGRPCClient("://bad-url")
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}
