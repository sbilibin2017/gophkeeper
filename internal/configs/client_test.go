package configs

import (
	"errors"
	"net"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Тест GetServerType для разных конфигураций
func TestGetServerType(t *testing.T) {
	cfg := &ClientConfig{}
	assert.Equal(t, "", cfg.GetServerType())

	cfg.HTTPClient = resty.New()
	assert.Equal(t, ServerTypeHTTP, cfg.GetServerType())

	cfg.HTTPClient = nil
	cfg.GRPCClient = &grpc.ClientConn{} // исправлено: используем реальный тип
	assert.Equal(t, ServerTypeGRPC, cfg.GetServerType())
}

// Тест WithDB — проверяем успешное и неуспешное подключение
func TestWithDB(t *testing.T) {
	// Ошибка при пустом пути
	opt := WithDB()
	cfg := &ClientConfig{}
	err := opt(cfg)
	require.Error(t, err)

	// Успешное подключение с sqlite in-memory
	opt = WithDB(":memory:")
	cfg = &ClientConfig{}
	err = opt(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.DB)

	// Проверка, что db.Ping действительно вызывается
	err = cfg.DB.Ping()
	require.NoError(t, err)
}

// Тест WithHTTPClient
func TestWithHTTPClient(t *testing.T) {
	// Ошибка если пустой URL
	opt := WithHTTPClient()
	cfg := &ClientConfig{}
	err := opt(cfg)
	require.Error(t, err)

	// Ошибка если неправильный URL
	opt = WithHTTPClient("://bad-url")
	cfg = &ClientConfig{}
	err = opt(cfg)
	require.Error(t, err)

	// Успешная настройка
	opt = WithHTTPClient("http://localhost:8080")
	cfg = &ClientConfig{}
	err = opt(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.HTTPClient)
	assert.Equal(t, "http://localhost:8080", cfg.HTTPClient.BaseURL)
}

// Простая реализация сервера для теста
type testServer struct{}

func startTestGRPCServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0") // слушаем случайный свободный порт
	require.NoError(t, err)

	s := grpc.NewServer()

	go func() {
		_ = s.Serve(lis)
	}()

	// Возвращаем адрес и функцию для остановки сервера
	return lis.Addr().String(), func() {
		s.Stop()
		lis.Close()
	}
}

func TestWithGRPCClient(t *testing.T) {
	// Запускаем тестовый сервер
	addr, stopServer := startTestGRPCServer(t)
	defer stopServer()

	cfg := &ClientConfig{}

	opt := WithGRPCClient(addr)
	err := opt(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.GRPCClient)

	// Проверяем, что соединение активно — вызываем State()
	state := cfg.GRPCClient.GetState()
	require.NotEqual(t, connectivity.Shutdown, state)

	// Закрываем соединение
	err = cfg.GRPCClient.Close()
	require.NoError(t, err)
}

// Тест WithJWTGenerator
func TestWithJWTGenerator(t *testing.T) {
	// Ошибка если пустой секрет
	opt := WithJWTGenerator()
	cfg := &ClientConfig{}
	err := opt(cfg)
	require.Error(t, err)

	// Успешное создание генератора
	opt = WithJWTGenerator("secretkey")
	cfg = &ClientConfig{}
	err = opt(cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.JWTGenerator)

	token, err := cfg.JWTGenerator("username")
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// Тест SetServerURLToEnv и GetServerURLFromEnv
func TestSetAndGetServerURLFromEnv(t *testing.T) {
	orig := os.Getenv("GOPHKEEPER_SERVER_URL")
	defer os.Setenv("GOPHKEEPER_SERVER_URL", orig)

	url := "http://example.com/api"
	err := SetServerURLToEnv(url)
	require.NoError(t, err)

	got := GetServerURLFromEnv()
	assert.Equal(t, url, got)

	err = os.Unsetenv("GOPHKEEPER_SERVER_URL")
	require.NoError(t, err)
	assert.Empty(t, GetServerURLFromEnv())
}

// Тест SetTokenToEnv и GetTokenFromEnv
func TestSetAndGetTokenFromEnv(t *testing.T) {
	orig := os.Getenv("GOPHKEEPER_TOKEN")
	defer os.Setenv("GOPHKEEPER_TOKEN", orig)

	token := "mysecrettoken"
	err := SetTokenToEnv(token)
	require.NoError(t, err)

	got := GetTokenFromEnv()
	assert.Equal(t, token, got)

	err = os.Unsetenv("GOPHKEEPER_TOKEN")
	require.NoError(t, err)
	assert.Empty(t, GetTokenFromEnv())
}

func TestNewClientConfig(t *testing.T) {
	// Опция, которая успешно настраивает HTTP клиент
	httpOpt := func(c *ClientConfig) error {
		c.HTTPClient = resty.New()
		return nil
	}

	// Опция, которая всегда возвращает ошибку
	errOpt := func(c *ClientConfig) error {
		return errors.New("test error")
	}

	// Успешный случай: опция не возвращает ошибку
	cfg, err := NewClientConfig(httpOpt)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.HTTPClient)

	// Ошибка: одна из опций возвращает ошибку
	cfg, err = NewClientConfig(httpOpt, errOpt)
	require.Error(t, err)
	require.Nil(t, cfg)
}
