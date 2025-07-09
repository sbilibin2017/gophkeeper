package configs

import (
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestNewClientConfig_Success(t *testing.T) {
	cfg, err := NewClientConfig(
		WithDB(),
		WithHTTPClient("http://localhost"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.DB)
	assert.NotNil(t, cfg.HTTPClient)

	assert.Nil(t, cfg.GRPCClient)

	assert.IsType(t, &sqlx.DB{}, cfg.DB)
	assert.IsType(t, &resty.Client{}, cfg.HTTPClient)

	// Закрываем базу
	if cfg.DB != nil {
		_ = cfg.DB.Close()
	}

	// Удаляем файл db.sqlite из текущей директории
	_ = os.Remove("db.sqlite")
}

func TestWithHTTPClient_InvalidURL(t *testing.T) {
	cfg := &ClientConfig{}
	err := WithHTTPClient("://bad-url")(cfg)
	assert.Error(t, err)
	assert.Nil(t, cfg.HTTPClient)
}

func TestWithGRPCClient_InvalidURL(t *testing.T) {
	cfg := &ClientConfig{}
	err := WithGRPCClient("://bad-url")(cfg)
	assert.Error(t, err)
	assert.Nil(t, cfg.GRPCClient)
}

func TestNewClientConfig_ReturnsFirstError(t *testing.T) {
	cfg, err := NewClientConfig(
		WithHTTPClient("://bad-url"),
	)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestWithGRPCClient(t *testing.T) {
	t.Run("valid grpc url", func(t *testing.T) {
		cfg := &ClientConfig{}
		// Для теста можно использовать любой валидный URL, например localhost с портом.
		err := WithGRPCClient("dns:///localhost:50051")(cfg)

		// Ошибок быть не должно, grpc клиент должен быть установлен
		assert.NoError(t, err)
		assert.NotNil(t, cfg.GRPCClient)
	})

	t.Run("invalid grpc url", func(t *testing.T) {
		cfg := &ClientConfig{}
		err := WithGRPCClient("://bad-url")(cfg)

		assert.Error(t, err)
		assert.Nil(t, cfg.GRPCClient)
	})
}

func TestWithToken(t *testing.T) {
	t.Run("empty token", func(t *testing.T) {
		cfg := &ClientConfig{}
		err := WithToken("")(cfg)
		assert.NoError(t, err)
		assert.Empty(t, cfg.Token)
	})

	t.Run("non-empty token", func(t *testing.T) {
		cfg := &ClientConfig{}
		token := "some_jwt_token"
		err := WithToken(token)(cfg)
		assert.NoError(t, err)
		assert.Equal(t, token, cfg.Token)
	})
}
